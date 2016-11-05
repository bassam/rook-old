package cephmgr

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"

	"github.com/rook/rook/pkg/cephmgr/client"
	"github.com/rook/rook/pkg/clusterd"
	"github.com/rook/rook/pkg/util"
	"github.com/rook/rook/pkg/util/proc"
)

const (
	monitorAgentName = "monitor"
)

type monAgent struct {
	factory client.ConnectionFactory
	monProc *proc.MonitoredProc
}

func (a *monAgent) Name() string {
	return monitorAgentName
}

func (a *monAgent) Initialize(context *clusterd.Context) error {
	return nil
}

func (a *monAgent) ConfigureLocalService(context *clusterd.Context) error {
	// check if the monitor is in the desired state for this node
	key := path.Join(cephKey, monitorAgentName, desiredKey, context.NodeID)
	monDesired, err := util.EtcdDirExists(context.EtcdClient, key)
	if err != nil {
		return err
	}
	if !monDesired {
		return a.DestroyLocalService(context)
	}

	cluster, err := LoadClusterInfo(context.EtcdClient)
	if err != nil {
		return fmt.Errorf("failed to load cluster info: %+v", err)
	}

	var ok bool
	var monitor *CephMonitorConfig
	if monitor, ok = cluster.Monitors[context.NodeID]; !ok {
		return fmt.Errorf("failed to find monitor during config")
	}

	// initialze the file system for the monitor
	if err := a.makeMonitorFileSystem(context, cluster, monitor.Name); err != nil {
		return fmt.Errorf("failed to make monitor filesystems: %+v", err)
	}

	// run the monitor
	err = a.runMonitor(context, cluster, monitor)
	if err != nil {
		return fmt.Errorf("failed to run monitors: %+v", err)
	}

	log.Printf("successfully started monitor %s", monitor.Name)

	return err
}

// stops and removes the monitor from this node
func (a *monAgent) DestroyLocalService(context *clusterd.Context) error {
	if a.monProc == nil {
		log.Printf("no need to stop a monitor that is not running")
		return nil
	}

	if err := a.monProc.Stop(); err != nil {
		log.Printf("failed to stop mon. %v", err)
		return err
	}

	log.Printf("stopped ceph monitor")
	a.monProc = nil

	// TODO: Clean up the monitor folder

	return nil
}

// creates and initializes the given monitors file systems
func (a *monAgent) makeMonitorFileSystem(context *clusterd.Context, cluster *ClusterInfo, monName string) error {
	// write the keyring to disk
	if err := writeMonitorKeyring(context.ConfigDir, monName, cluster); err != nil {
		return err
	}

	// write the config file to disk
	confFilePath, err := generateConnectionConfigFile(context, cluster, getMonRunDirPath(context.ConfigDir, monName),
		"admin", getMonKeyringPath(context.ConfigDir, monName), context.Debug)
	if err != nil {
		return err
	}

	// create monitor data dir
	monDataDir := getMonDataDirPath(context.ConfigDir, monName)
	if err := os.MkdirAll(monDataDir, 0744); err != nil {
		fmt.Printf("failed to create monitor data directory at %s: %+v", monDataDir, err)
	}

	// write the kv_backend file to force ceph to use rocksdb for the MON store
	if err := writeBackendFile(monDataDir, "rocksdb"); err != nil {
		return err
	}

	// call mon --mkfs in a child process
	err = context.ProcMan.Run(
		"mon",
		"--mkfs",
		fmt.Sprintf("--cluster=%s", cluster.Name),
		fmt.Sprintf("--name=mon.%s", monName),
		fmt.Sprintf("--mon-data=%s", monDataDir),
		fmt.Sprintf("--conf=%s", confFilePath),
		fmt.Sprintf("--keyring=%s", getMonKeyringPath(context.ConfigDir, monName)))
	if err != nil {
		return fmt.Errorf("failed mon %s --mkfs: %+v", monName, err)
	}

	return nil
}

// runs the monitor in a child process
func (a *monAgent) runMonitor(context *clusterd.Context, cluster *ClusterInfo, monitor *CephMonitorConfig) error {
	if monitor.Endpoint == "" {
		return fmt.Errorf("missing endpoint for mon %s", monitor.Name)
	}

	// start the monitor daemon in the foreground with the given config
	log.Printf("starting monitor %s", monitor.Name)
	monNameArg := fmt.Sprintf("--name=mon.%s", monitor.Name)
	monProc, err := context.ProcMan.Start(
		"mon",
		regexp.QuoteMeta(monNameArg),
		proc.ReuseExisting,
		"--foreground",
		fmt.Sprintf("--cluster=%s", cluster.Name),
		monNameArg,
		fmt.Sprintf("--mon-data=%s", getMonDataDirPath(context.ConfigDir, monitor.Name)),
		fmt.Sprintf("--conf=%s", getConfFilePath(getMonRunDirPath(context.ConfigDir, monitor.Name), cluster.Name)),
		fmt.Sprintf("--public-addr=%s", monitor.Endpoint))
	if err != nil {
		return fmt.Errorf("failed to start monitor %s: %v", monitor.Name, err)
	}

	if monProc != nil {
		// if the process was already running Start will return nil in which case we don't want to overwrite it
		a.monProc = monProc
	}

	return nil
}

// writes the monitor keyring to disk
func writeMonitorKeyring(configDir, monName string, c *ClusterInfo) error {
	keyring := fmt.Sprintf(monitorKeyringTemplate, c.MonitorSecret, c.AdminSecret)
	keyringPath := getMonKeyringPath(configDir, monName)
	if err := os.MkdirAll(filepath.Dir(keyringPath), 0744); err != nil {
		return fmt.Errorf("failed to create keyring directory for %s: %+v", keyringPath, err)
	}
	if err := ioutil.WriteFile(keyringPath, []byte(keyring), 0644); err != nil {
		return fmt.Errorf("failed to write monitor keyring to %s: %+v", keyringPath, err)
	}

	return nil
}

// writes the monitor backend file to disk
func writeBackendFile(monDataDir, backend string) error {
	backendPath := filepath.Join(monDataDir, "kv_backend")
	if err := ioutil.WriteFile(backendPath, []byte(backend), 0644); err != nil {
		return fmt.Errorf("failed to write kv_backend to %s: %+v", backendPath, err)
	}
	return nil
}
