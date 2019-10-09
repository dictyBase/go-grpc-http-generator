# go-grpc-http-generator
[![License](https://img.shields.io/badge/License-BSD%202--Clause-blue.svg)](LICENSE)   
[![Go Report Card](https://goreportcard.com/badge/github.com/dictyBase/go-grpc-http-generator)](https://goreportcard.com/report/github.com/dictyBase/go-grpc-http-generator)
[![Technical debt](https://badgen.net/codeclimate/tech-debt/dictyBase/go-grpc-http-generator)](https://codeclimate.com/github/dictyBase/go-grpc-http-generator/trends/technical_debt)
[![Issues](https://badgen.net/codeclimate/issues/dictyBase/go-grpc-http-generator)](https://codeclimate.com/github/dictyBase/go-grpc-http-generator/issues)
[![Maintainability](https://api.codeclimate.com/v1/badges/fa67b8ce344f4c7bf7b1/maintainability)](https://codeclimate.com/github/dictyBase/go-grpc-http-generator/maintainability)
[![Dependabot Status](https://api.dependabot.com/badges/status?host=github&repo=dictyBase/go-grpc-http-generator)](https://dependabot.com)   
![GitHub repo size](https://img.shields.io/github/repo-size/dictyBase/go-grpc-http-generator?style=plastic)
![GitHub code size in bytes](https://img.shields.io/github/languages/code-size/dictyBase/go-grpc-http-generator?style=plastic)
[![Lines of Code](https://badgen.net/codeclimate/loc/dictyBase/go-grpc-http-generator)](https://codeclimate.com/github/dictyBase/go-grpc-http-generator/code)   
![Commits](https://badgen.net/github/commits/dictyBase/go-grpc-http-generator/master)
![Last commit](https://badgen.net/github/last-commit/dictyBase/go-grpc-http-generator/master)
![Branches](https://badgen.net/github/branches/dictyBase/go-grpc-http-generator)
![Tags](https://badgen.net/github/tags/dictyBase/go-grpc-http-generator)   
![Issues](https://badgen.net/github/issues/dictyBase/go-grpc-http-generator)
![Open Issues](https://badgen.net/github/open-issues/dictyBase/go-grpc-http-generator)
![Closed Issues](https://badgen.net/github/closed-issues/dictyBase/go-grpc-http-generator)   
![Total PRS](https://badgen.net/github/prs/dictyBase/go-grpc-http-generator)
![Open PRS](https://badgen.net/github/open-prs/dictyBase/go-grpc-http-generator)
![Closed PRS](https://badgen.net/github/closed-prs/dictyBase/go-grpc-http-generator)
![Merged PRS](https://badgen.net/github/merged-prs/dictyBase/go-grpc-http-generator)   
[![Funding](https://badgen.net/badge/NIGMS/Rex%20L%20Chisholm,dictyBase/yellow?list=|)](https://projectreporter.nih.gov/project_info_description.cfm?aid=9476993)
[![Funding](https://badgen.net/badge/NIGMS/Rex%20L%20Chisholm,DSC/yellow?list=|)](https://projectreporter.nih.gov/project_info_description.cfm?aid=9438930)

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

## Developers
<a href="https://sourcerer.io/cybersiddhu"><img src="https://sourcerer.io/assets/avatar/cybersiddhu" height="80px" alt="Sourcerer"></a>
<a href="https://sourcerer.io/wildlifehexagon"><img src="https://sourcerer.io/assets/avatar/wildlifehexagon" height="80px" alt="Sourcerer"></a>
