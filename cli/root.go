// Copyright 2016,2017 Yaacov Zamir <kobi.zamir@gmail.com>
// and other contributors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package cli command line interface
package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/MohawkTSDB/mohawk/server"
)

// AUTHOR the author name and Email
const AUTHOR = "Yaacov Zamir <kobi.zamir@gmail.com>"

// defaults
const defaultTLSKey = "server.key"
const defaultTLSCert = "server.pem"

// RootCmd Mohawk root cli Command
var RootCmd = &cobra.Command{
	Use: "mohawk",
	Long: fmt.Sprintf(`Mohawk is a metric data storage engine.

Mohawk is a metric data storage engine that uses a plugin architecture for data
storage and a simple REST API as the primary interface.

Version:
  %s

Author:
  %s`, api.VER, AUTHOR),
	Run: func(cmd *cobra.Command, args []string) {

		// Print version and quit
		if viper.GetBool("version") {
			fmt.Printf("Mohawk version: %s\n\n", api.VER)
			return
		}

		// Run the REST API server
		api.Serve()
	},
}

func init() {
	cobra.OnInitialize(initConfig)

	// Flag definition
	RootCmd.Flags().StringP("storage", "b", "memory", "the storage plugin to use")
	RootCmd.Flags().String("token", "", "authorization token")
	RootCmd.Flags().String("media", "./mohawk-webui", "path to media files")
	RootCmd.Flags().String("key", defaultTLSKey, "path to TLS key file")
	RootCmd.Flags().String("cert", defaultTLSCert, "path to TLS cert file")
	RootCmd.Flags().String("options", "", "specific storage options [e.g. db-dirname, db-url]")
	RootCmd.Flags().IntP("port", "p", 8080, "server port")
	RootCmd.Flags().BoolP("tls", "t", false, "use TLS server")
	RootCmd.Flags().BoolP("gzip", "g", false, "use gzip encoding")
	RootCmd.Flags().BoolP("verbose", "V", false, "more debug output")
	RootCmd.Flags().BoolP("version", "v", false, "display mohawk version number")
	RootCmd.Flags().StringP("config", "c", "", "config file")

	// Viper Binding
	viper.BindPFlag("storage", RootCmd.Flags().Lookup("storage"))
	viper.BindPFlag("token", RootCmd.Flags().Lookup("token"))
	viper.BindPFlag("media", RootCmd.Flags().Lookup("media"))
	viper.BindPFlag("key", RootCmd.Flags().Lookup("key"))
	viper.BindPFlag("cert", RootCmd.Flags().Lookup("cert"))
	viper.BindPFlag("options", RootCmd.Flags().Lookup("options"))
	viper.BindPFlag("port", RootCmd.Flags().Lookup("port"))
	viper.BindPFlag("tls", RootCmd.Flags().Lookup("tls"))
	viper.BindPFlag("gzip", RootCmd.Flags().Lookup("gzip"))
	viper.BindPFlag("verbose", RootCmd.Flags().Lookup("verbose"))
	viper.BindPFlag("version", RootCmd.Flags().Lookup("version"))
	viper.BindPFlag("config", RootCmd.Flags().Lookup("config"))
}

func initConfig() {
	if viper.GetString("config") != "" {
		viper.SetConfigFile(viper.GetString("config"))
		if err := viper.ReadInConfig(); err != nil {
			fmt.Println("Error reading config file:", err)
		}
	}
}
