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
package flags

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func VerifyRequiredFlags(cmd *cobra.Command, requiredFlags []string) error {
	var missingFlags []string
	for _, reqFlag := range requiredFlags {
		val, err := cmd.Flags().GetString(reqFlag)
		if err != nil || val == "" {
			missingFlags = append(missingFlags, reqFlag)
		}
	}

	return createRequiredFlagError(cmd.Name(), missingFlags)
}

func VerifyRequiredUint64Flags(cmd *cobra.Command, requiredFlags []string) error {
	var missingFlags []string
	for _, reqFlag := range requiredFlags {
		val, err := cmd.Flags().GetUint64(reqFlag)
		if err != nil || val == 0 {
			missingFlags = append(missingFlags, reqFlag)
		}
	}

	return createRequiredFlagError(cmd.Name(), missingFlags)
}

func createRequiredFlagError(name string, flags []string) error {
	if len(flags) == 0 {
		return nil
	}

	if len(flags) == 1 {
		return fmt.Errorf("%s is required for %s", flags[0], name)
	}

	return fmt.Errorf("%s are required for %s", strings.Join(flags, ","), name)
}
