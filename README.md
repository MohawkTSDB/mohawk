# mohawk

MOck HAWKular is a mock Hawkular server for testing.

## Introduction

Utility server for testing Hawkular clients, the server can mock
a running metrics Hawkular server. It can use different backends for different test use cases.

  - Random backend, mimics lots of metrics available only for reading.
  - Sqlite backend, mimics persistable read and write.


## License and copyright

```
   Copyright 2016 Red Hat, Inc. and/or its affiliates
   and other contributors.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
```

## Installation

To install, get the source code, or do ``go install github.com/yaacov/mohawk`` if using go.
To run, users will need the ``server.key`` and ``server.pem`` certification files.

Or if using a Copr repository for Fedora:

```
sudo dnf copr enable yaacov/mohawk 
sudo dnf install mohawk
```

## Mock Certifications

The server use mock certification to serve ``https`` requests. Users can use mock 
self signed credentials files for this. 

To create a self signed credentials use this bash commands:
```
openssl ecparam -genkey -name secp384r1 -out server.key
openssl req -new -x509 -sha256 -key server.key -out server.pem -days 3650
```

## Usage

When installed, run using the command line ``mohawk``, when run from code, use ``go run *.go``
 
The `-h` flag will print out a help text, that list the command line arguments.

```bash
# run `go run *.go` from the source directory, or if installed use:
mohawk -h

Usage of mohawk:
  -api string
    	the hawkulr api to mimic [e.g. 0.8.9.Testing, 0.21.2.Final] (default "0.21.0")
  -backend string
    	the backend to use [random, sqlite] (default "random")
  -port int
    	server port (default 8443)
  -tls string
    	use TLS server (default "true")
```

## Example of use

Running from the source directory using ``go run`` and requesting the help message.

```bash
go run *.go -h
Usage of mohawk:
...
```
Running from system install using ``mohawk`` without ``tls`` and using the ``sqlite`` back end.

```bash
mohawk -tls false -port 8080 -backend sqlite
2017/01/03 10:06:50 Start server, listen on http://0.0.0.0:8080
```

Running from the source directory using ``go run`` and the ``sqlite`` back end
[ Remmeber to set up the ``server.key`` and ``server.pem`` files in your diretory ].

```bash
go run *.go -backend sqlite -port 8443
2016/12/01 14:23:48 Start server, listen on https://0.0.0.0:8443
```
