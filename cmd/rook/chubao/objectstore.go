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

package chubao

import (
	"github.com/rook/rook/cmd/rook/rook"
	"github.com/rook/rook/pkg/clusterd"
	"github.com/rook/rook/pkg/operator/k8sutil"
	"github.com/rook/rook/pkg/util/flags"
	"github.com/spf13/cobra"
)

var objectStoreCmd = &cobra.Command{
	Use:   "objectstore",
	Short: "Deploys the chubao ObjectNode for supporting s3 service",
}

func init() {
	flags.SetFlagsFromEnv(objectStoreCmd.Flags(), rook.RookEnvVarPrefix)
	flags.SetLoggingFlags(objectStoreCmd.Flags())
	objectStoreCmd.RunE = startObjectStore
}

func startObjectStore(cmd *cobra.Command, args []string) error {
	rook.SetLogLevel()
	rook.LogStartupInfo(objectStoreCmd.Flags())

	//logger.Info("starting Rook-Chubao operator")
	context := createContext()
	context.NetworkInfo = clusterd.NetworkInfo{}
	context.ConfigDir = k8sutil.DataDir
	return nil
}
