// Copyright 2016,2017,2018 Yaacov Zamir <kobi.zamir@gmail.com>
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
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/MohawkTSDB/mohawk/src/server"
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
  %s`, server.VER, AUTHOR),
	Run: func(cmd *cobra.Command, args []string) {

		// Print version and quit
		if viper.GetBool("version") {
			fmt.Printf("Mohawk version: %s\n\n", server.VER)
			return
		}

		// Run the REST API server
		server.Serve()
	},
}

func init() {
	cobra.OnInitialize(initConfig)

	// Flag definition
	RootCmd.Flags().StringP("storage", "b", "memory", "the storage plugin to use")
	RootCmd.Flags().String("bearer-auth", "", "token used for bearer authorization")
	RootCmd.Flags().String("basic-auth", "", "authorization user and password pair (e.g. jack:secret-password)")
	RootCmd.Flags().String("media", "./mohawk-webui", "path to media files")
	RootCmd.Flags().String("key", defaultTLSKey, "path to TLS key file")
	RootCmd.Flags().String("cert", defaultTLSCert, "path to TLS cert file")
	RootCmd.Flags().String("options", "", "specific storage options, use \"--options=help\" for help")
	RootCmd.Flags().IntP("port", "p", 8080, "server port")
	RootCmd.Flags().BoolP("tls", "t", false, "use TLS server")
	RootCmd.Flags().BoolP("gzip", "g", false, "use gzip encoding")
	RootCmd.Flags().BoolP("verbose", "V", false, "more debug output")
	RootCmd.Flags().BoolP("version", "v", false, "display mohawk version number")
	RootCmd.Flags().StringP("config", "c", "", "config file")
	RootCmd.Flags().Int("alerts-interval", 5, "Check alerts every N sec")
	RootCmd.Flags().String("alerts-server", "http://localhost:9099/append", "Alert buffer URL")
	RootCmd.Flags().String("alerts-server-method", "POST", "Alert server http method")
	RootCmd.Flags().Bool("alerts-server-insecure", false, "Alert server https skip verify")
	RootCmd.Flags().String("default-tenant", "_ops", "Default tenant to use")

	// Viper Binding
	viper.BindPFlag("storage", RootCmd.Flags().Lookup("storage"))
	viper.BindPFlag("bearer-auth", RootCmd.Flags().Lookup("bearer-auth"))
	viper.BindPFlag("basic-auth", RootCmd.Flags().Lookup("basic-auth"))
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
	viper.BindPFlag("alerts-interval", RootCmd.Flags().Lookup("alerts-interval"))
	viper.BindPFlag("alerts-server", RootCmd.Flags().Lookup("alerts-server"))
	viper.BindPFlag("alerts-server-method", RootCmd.Flags().Lookup("alerts-server-method"))
	viper.BindPFlag("alerts-server-insecure", RootCmd.Flags().Lookup("alerts-server"))
	viper.BindPFlag("default-tenant", RootCmd.Flags().Lookup("default-tenant"))
}

func initConfig() {
	if viper.GetString("config") != "" {
		viper.SetConfigFile(viper.GetString("config"))
		if err := viper.ReadInConfig(); err != nil {
			log.Println("Error reading config file:", err)
		}
	}
}
