// +build linux,amd64 linux,arm64

/*
Copyright 2016 The Rook Authors. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/rook/rook/pkg/cephmgr/cephd"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of rookd",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("cephd: %v\n", cephd.Version())
		rmajor, rminor, rpatch := cephd.RadosVersion()
		fmt.Printf("rados: %v.%v.%v\n", rmajor, rminor, rpatch)
		return nil
	},
}
