# go-grpc-http-generator
A golang cli to generate golang gRPC and HTTP-JSON reverse proxy code from
[dictybase api(proto3) definition
repository](https://github.com/dictyBase/dictybaseapis).

## Installation
```
go get -v -u github.com/dictyBase/go-grpc-http-generator
```

## Usage 
```
NAME:
   go-grpc-http-generator - cli for generating go gRPC and gRPC-gateway source code for dictybase api and services

USAGE:
   go-grpc-http-generator [global options] command [command options] [arguments...]

VERSION:
   1.0.0

COMMANDS:
     help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --output value, -o value  Base output path for generated source code.
                               By default, it is $GOPATH/src along with the appended package path
   --prefix value            Go package prefix that will be matched to select definition files for code generation (default: "github.com/dictyBase/go-genproto")
   --api-repo value          Repository containing protocol buffer definitions of google apis, will be check out or updated under GOPATH (default: "https://github.com/googleapis/googleapis")
   --proto-repo value        Repository containing core protocol buffer definitions from google, will be checked out or updated under GOPATH (default: "https://github.com/google/protobuf")
   --dictybase-repo value    Repository containing protocol buffer definitions of dictybase api and services, will be checked out or updated under GOPATH (default: "https://github.com/dictyBase/dictybaseapis")
   --log-level value         log level for the application (default: "error")
   --log-format value        format of the logging out, either of json or text (default: "text")
   --help, -h                show help
   --version, -v             print the version
```
