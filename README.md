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

Help extract:
```
tge-cli init creates a TGE workspace.

Usage:
    tge-cli init package

Package argument can be of several forms:
    local   ex: my-app
    url     ex: github.com/me/my-app

In both cases, the last token will be used as worspace root.
```

## Build the application
Once the application folder is created, releases can be generated using:
```shell
tge-cli build -target [target] [package-path]
```
Target allows to build yoour application for Desktop, Mobile or Web backend. See [tge-cli](https://github.com/thommil/tge-cli) for full details on how to use it.)

Targets details:
 * [Android](https://github.com/Thommil/tge/tree/master/template/android)
 * [Browser](https://github.com/Thommil/tge/tree/master/template/browser)
 * [IOS](https://github.com/Thommil/tge/tree/master/template/ios)
 * [Linux](https://github.com/Thommil/tge/tree/master/template/linux)
 * [MacOS](https://github.com/Thommil/tge/tree/master/template/darwin)
 * [Windows](https://github.com/Thommil/tge/tree/master/template/windows)

Help extract:
```
tge-cli build build and deploys TGE applications.

Usage:
    tge-cli build [-target TARGET] [-dev] [-v] [-bundleid ID] packagePath

The package path must point to a valid TGE application, the generated
application will be stored in the dist/$TARGET folder.

-target     defines the application target:
                desktop (default)
                browser
                android
                ios

            For desktop target, the generated application depends on current OS:
                MacOS   -> darwin
                Windows -> windows
                Linux   -> linux

            For each target, the corresponding folder in your workspace will contain
            additional ressources for more customization (see README.md files)

-dev        dev flag allows to generate application faster by omitting assets copy.
            Desktop applications are not packed and console remains opened.
                        On Android the resulting APK will support all architectures.
            Debug mode is also enabled.

-v          verbose output for debugging purpose

-bundleid  is mandatory for IOS build and can be obtained from Apple Developer.
```