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
   genproto - cli for generating go gRPC and gRPC-gateway source code for dictybase api and services

USAGE:
   go-grpc-http-generator [global options] command [command options] [arguments...]

VERSION:
   2.0.0

AUTHOR:
   Siddhartha Basu

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --output value, -o value        Base output path for generated source code.
                                     By default, it is $GOPATH/src along with the appended package path
   --prefix value                  Go package prefix that will be matched to select definition files for code generation (default: "github.com/dictyBase/go-genproto")
   --api-repo value                Repository containing protocol buffer definitions of google apis, will be check out (default: "https://github.com/googleapis/googleapis")
   --proto-repo value              Repository containing core protocol buffer definitions from google, will be checked out (default: "https://github.com/protocolbuffers/protobuf")
   --proto-repo-tag value          Repository tag for protocol buffer repo (default: "v3.9.2")
   --validator-repo value          Repository containing protocol buffer definitions for validation, will be checked out (default: "https://github.com/mwitkow/go-proto-validators.git")
   --validator-repo-tag value      Repository tag for validation protocol buffer (default: "v0.2.0")
   --input-folder value, -i value  Folder containing protocol buffer definitions, will be looked up recursively
   --log-level value               log level for the application (default: "error")
   --log-format value              format of the logging out, either of json or text (default: "json")
   --swagger-gen                   generate swagger definition files from grpc-gateway definition
   --swagger-output value          Output folder for swagger definition files, should be set with swagger-gen option
   --help, -h                      show help
   --version, -v                   print the version
```
