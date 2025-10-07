# Rubberband command-line interface

A rubberband command line client, written in go.

## Development

Read a bit about [go development](https://golang.org/doc/) if you're new to golang. If you have a go development environment already set up, then simply:

```
go get github.com/scipopt/rubberband-cli
```

or:
```
go mod tidy
go build .
```

## Use

```
$ rubberband-cli
NAME:
   rubberband-cli - the rubberband command line client

USAGE:
   rubberband-cli [global options] command [command options] [arguments...]
   
VERSION:
   1.0.0
   
COMMANDS:
     upload, up  Upload a list of related output files for parsing and processing.
     help, h     Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --async        Run this command asynchronously. Receive the results in an email.
   --tags value   A comma-separated list of tags to associate with the uploaded run.
   -e value       The expirationdate, should be provided in form "2017-Aug-24".
   --help, -h     show help
   --version, -v  print the version
```

Standard usage:

```
RUBBERBAND_URL="https://rubberband.example.com" RUBBERBAND_API_KEY="MYAPIKEY" rubberband-cli upload myfile.{out,set,err,meta}
```

The following environment variables are required to run:
- RUBBERBAND_URL
- RUBBERBAND_API_KEY

The following optional environment variable can also be set:
- RBCLI_USE_LDAP 
