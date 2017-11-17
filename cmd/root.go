package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"fmt"
	"net/http"
	"time"
	"log"
	"net/url"

	"github.com/MohawkTSDB/mohawk/middleware"
	"github.com/MohawkTSDB/mohawk/backend"
	"github.com/MohawkTSDB/mohawk/backend/sqlite"
	"github.com/MohawkTSDB/mohawk/backend/memory"
	"github.com/MohawkTSDB/mohawk/backend/mongo"
	"github.com/MohawkTSDB/mohawk/backend/example"
	"github.com/MohawkTSDB/mohawk/router"
)

// VER the server version
const VER = "0.21.4"

// defaults
const defaultAPI = "0.21.0"
const defaultTLSKey = "server.key"
const defaultTLSCert = "server.pem"

var RootCmd = &cobra.Command{
	Use:    "mohakwk",
	Short:  "Mohawk is a fast, lightweight time series database.",
	Run: func(cmd *cobra.Command, args []string) {
		serve()
	},
}

var BackendName string
var cfgFile string
var plugin string
var token string
var media string
var key string
var cert string
var options string
var port int
var tls bool
var gzip bool
var verbose bool

func init() {
	cobra.OnInitialize(initConfig)
	// Setting Defaults
	viper.SetDefault("plugin", "memory")
	viper.SetDefault("token", "")
	viper.SetDefault("media", "./mohawk-webui")
	viper.SetDefault("key", defaultTLSKey)
	viper.SetDefault("cert", defaultTLSCert)
	viper.SetDefault("options","")
	viper.SetDefault("port",8080)
	viper.SetDefault("tls",false)
	viper.SetDefault("gzip",false)
	viper.SetDefault("verbose", false)

	// Flag definition
	RootCmd.Flags().StringVar(&cfgFile, "config", "","config file (default is $HOME/.cobra.yaml)")
	RootCmd.Flags().StringVar(&plugin, "plugin","memory", "the backend driver to use")
	RootCmd.Flags().StringVar(&token, "token", "","authorization token")
	RootCmd.Flags().StringVar(&media, "media", "./mohawk-webui", "path to media files")
	RootCmd.Flags().StringVar(&key, "key", defaultTLSKey, "path to TLS key file")
	RootCmd.Flags().StringVar(&cert, "cert", defaultTLSCert, "path to TLS cert file")
	RootCmd.Flags().StringVar(&options, "options", "", "specific backend options [e.g. db-dirname, db-url]")
	RootCmd.Flags().IntVarP(&port, "port","p", 8080, "server port")
	RootCmd.Flags().BoolVarP(&tls, "tls","t",false, "use TLS server")
	RootCmd.Flags().BoolVarP(&gzip, "gzip","g",false, "use gzip encoding")
	RootCmd.Flags().BoolVarP(&verbose, "verbose","v", false, "more debug output")
}

func initConfig(){
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
		if err := viper.ReadInConfig(); err == nil {
			fmt.Println("Using config file:", viper.ConfigFileUsed()) // debug
			port = viper.GetInt("port")
			plugin = viper.GetString("plugin")
			key = viper.GetString("key")
			cert = viper.GetString("cert")
			tls = viper.GetBool("tls")
			gzip = viper.GetBool("gzip")
		}
	}
}

// GetStatus return a json status struct
func GetStatus(w http.ResponseWriter, r *http.Request, argv map[string]string) {
	resTemplate := `{"MetricsService":"STARTED","Implementation-Version":"%s","MohawkVersion":"%s","MohawkBackend":"%s"}`
	res := fmt.Sprintf(resTemplate, defaultAPI, VER, BackendName)

	w.WriteHeader(200)
	fmt.Fprintln(w, res)
}

func serve() error {
	var db backend.Backend
	var middlewareList []middleware.MiddleWare

	// Create and init the backend
	switch plugin {
	case "sqlite":
		db = &sqlite.Backend{}
	case "memory":
		db = &memory.Backend{}
	case "mongo":
		db = &mongo.Backend{}
	case "example":
		db = &example.Backend{}
	default:
		log.Fatal("Can't find backend:", plugin)
	}

	// parse options
	if options, err := url.ParseQuery(options); err == nil {
		db.Open(options)
	} else {
		log.Fatal("Can't parse opetions:", options)
	}

	// set global variables
	BackendName = db.Name()

	// h common variables to be used for the backend Handler functions
	// backend the backend to use for metrics source
	h := backend.Handler{
		Verbose: verbose,
		Backend: db,
	}

	// Create the routers
	// Requests not handled by the routers will be forworded to BadRequest Handler
	rRoot := router.Router{
		Prefix: "/hawkular/metrics/",
	}
	// Root Routing table
	rRoot.Add("GET", "status", GetStatus)
	rRoot.Add("GET", "tenants", h.GetTenants)
	rRoot.Add("GET", "metrics", h.GetMetrics)

	// Metrics Routing tables
	rGauges := router.Router{
		Prefix: "/hawkular/metrics/gauges/",
	}
	rGauges.Add("GET", ":id/raw", h.GetData)
	rGauges.Add("GET", ":id/stats", h.GetData)
	rGauges.Add("POST", "raw", h.PostData)
	rGauges.Add("POST", "raw/query", h.PostQuery)
	rGauges.Add("PUT", "tags", h.PutMultiTags)
	rGauges.Add("PUT", ":id/tags", h.PutTags)
	rGauges.Add("DELETE", ":id/raw", h.DeleteData)
	rGauges.Add("DELETE", ":id/tags/:tags", h.DeleteTags)

	// deprecated
	rGauges.Add("GET", ":id/data", h.GetData)
	rGauges.Add("POST", "data", h.PostData)
	rGauges.Add("POST", "stats/query", h.PostQuery)

	rCounters := router.Router{
		Prefix: "/hawkular/metrics/counters/",
	}
	rCounters.Add("GET", ":id/raw", h.GetData)
	rCounters.Add("GET", ":id/stats", h.GetData)
	rCounters.Add("POST", "raw", h.PostData)
	rCounters.Add("POST", "raw/query", h.PostQuery)
	rCounters.Add("PUT", ":id/tags", h.PutTags)

	// deprecated
	rCounters.Add("GET", ":id/data", h.GetData)
	rCounters.Add("POST", "data", h.PostData)
	rCounters.Add("POST", "stats/query", h.PostQuery)

	rAvailability := router.Router{
		Prefix: "/hawkular/metrics/availability/",
	}
	rAvailability.Add("GET", ":id/raw", h.GetData)
	rAvailability.Add("GET", ":id/stats", h.GetData)

	// Create the middlewares
	// logger a logging middleware
	logger := middleware.Logger{
		Verbose: verbose,
	}

	// static a file server middleware
	static := middleware.Static{
		Verbose:   verbose,
		MediaPath: media,
	}

	// authorization middleware
	authorization := middleware.Authorization{
		Verbose:         verbose,
		UseToken:        token != "",
		PublicPathRegex: "^/hawkular/metrics/status$",
		Token:           token,
	}

	// headers a headers middleware
	headers := middleware.Headers{
		Verbose: verbose,
	}

	// gzipper a gzip encoding middleware
	gzipper := middleware.GZipper{
		UseGzip: gzip,
		Verbose: verbose,
	}

	// badrequest a BadRequest middleware
	badrequest := middleware.BadRequest{
		Verbose: verbose,
	}

	// concat middlewars and routes (first logger until rRoot) with a fallback to BadRequest
	middlewareList = []middleware.MiddleWare{
		&logger,
		&authorization,
		&headers,
		&gzipper,
		&rGauges,
		&rCounters,
		&rAvailability,
		&rRoot,
		&static,
		&badrequest,
	}
	middleware.Append(middlewareList)

	// Run the server
	srv := &http.Server{
		Addr:           fmt.Sprintf("0.0.0.0:%d", port),
		Handler:        logger,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	if tls {
		log.Printf("Start server, listen on https://%+v", srv.Addr)
		log.Fatal(srv.ListenAndServeTLS(cert, key))
	} else {
		log.Printf("Start server, listen on http://%+v", srv.Addr)
		log.Fatal(srv.ListenAndServe())
	}

	return nil
}
