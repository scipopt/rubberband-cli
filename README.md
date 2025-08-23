# Rubberband command-line interface

A rubberband command line client, written in go.

## Development

Read a bit about [go development](https://golang.org/doc/) if you're new to golang. If you have a go development environment already set up, then simply:

```
go get github.com/scipopt/rubberband-cli
```

or:
```
go mod init rbcli
go get github.com/go-ldap/ldap/v3
go get github.com/urfave/cli
go build .
```

## Use

```
$ rbcli
NAME:
   rbcli - the rubberband command line client

USAGE:
   rbcli [global options] command [command options] [arguments...]
   
VERSION:
   0.0.3
   
COMMANDS:
     upload, up  Upload a list of related output files for parsing and processing.
     help, h     Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --async        Run this command asynchronously. Receive the results in an email.
   --tags value   A comma-separated list of tags to associate with the uploaded run.
   --help, -h     show help
   --version, -v  print the version
```

Standard usage:

```
rbcli upload myfile.{out,set,err,meta}
```
