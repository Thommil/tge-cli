## Install tge-cli
To get the client, run:
```shell
go get github.com/thommil/tge-cli
```

The command line tool should be available in the GOPATH/bin folder.

## Create new application
The create a new application workspace, run:
```shell
tge-cli init [package-name]
```

Ideally, the package-name should be based on standard Go package rules (ex: gihtub.com/thommil/my-app) but local package also works (ex: my-app).

An application folder will be created with all needed resources to begin. See [Go Doc](https://godoc.org/github.com/thommil/tge) for details.

## Build the application
Once the application folder is created, releases can be generated using:
```shell
tge-cli build -target [target] [package-path]
```
Target allows to build yoour application for Desktop, Mobile or Web backend. See [tge-cli](https://github.com/thommil/tge-cli) for full details on how to use it.)