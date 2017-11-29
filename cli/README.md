

# mohawk/usage

![Mohawk](/images/logo-128.png?raw=true "Mohawk Logo")

Mohawk is a metric data storage engine that uses a plugin architecture for data storage and a simple REST API as the primary interface.

## Usage

When installed, run using the command line ``mohawk``

```bash
mohawk --version

Mohawk version: 0.22.1
```

The `-h` flag will print out a help text, that list the command line arguments.

```
mohawk -h

Mohawk is a metric data storage engine.

Mohawk is a metric data storage engine that uses a plugin architecture for data
storage and a simple REST API as the primary interface.

Version:
  0.22.1

Author:
  Yaacov Zamir <kobi.zamir@gmail.com>

Usage:
  mohawk [flags]

Flags:
      --cert string      path to TLS cert file (default "server.pem")
  -c, --config string    config file
  -g, --gzip             use gzip encoding
  -h, --help             help for mohawk
      --key string       path to TLS key file (default "server.key")
      --media string     path to media files (default "./mohawk-webui")
      --options string   specific storage options [e.g. db-dirname, db-url]
  -p, --port int         server port (default 8080)
  -b, --storage string   the storage plugin to use (default "memory")
  -t, --tls              use TLS server
      --token string     authorization token
  -V, --verbose          more debug output
  -v, --version          display mohawk version number
```

Running ``mohawk`` with ``tls`` and using the ``memory`` back end.

```
mohawk --tls --gzip --port 8443 --storage memory
2017/06/30 11:37:08 Start server, listen on https://0.0.0.0:8443
```

###### When running with tls on, we need .key and .pem files:

This commands will crate self signed secrets for testing.

```
openssl ecparam -genkey -name secp384r1 -out server.key
openssl req -new -x509 -sha256 -key server.key -out server.pem -days 3650
```
