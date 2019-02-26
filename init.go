package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/otiai10/copy"
)

func (builder *Builder) initWorkspace(packageArg string) error {
	builder.packageName = packageArg
	if index := strings.LastIndex(builder.packageName, "/"); index >= 0 {
		builder.packagePath = path.Join(builder.cwd, builder.packageName[index:])
	} else {
		builder.packagePath = path.Join(builder.cwd, builder.packageName)
	}

	if _, err := os.Stat(builder.packagePath); os.IsNotExist(err) {
		if err = os.MkdirAll(builder.packagePath, os.ModeDir|os.FileMode(0777)); err != nil {
			return err
		}
	} else {
		log("ERROR", fmt.Sprintf("path %s already exists", builder.packagePath))
		os.Exit(2)
	}

	if err := os.Chdir(builder.packagePath); err != nil {
		return err
	}

	if err := builder.installTGE(); err != nil {
		return err
	}

	log("NOTICE", "Initializing project files")
	if err := copy.Copy(path.Join(builder.tgeRootPath, tgeTemplatePath), builder.packagePath); err != nil {
		log("ERROR", err.Error())
		return fmt.Errorf("Failed to copy project files, try manually from '%s", path.Join(builder.tgeRootPath, tgeTemplatePath))
	}

	return nil
}

func (builder *Builder) cleanInitBuilder() {
	os.RemoveAll(builder.packagePath)
}

func doInit(builder Builder) {
	os.Args = os.Args[1:]
	flag.Usage = func() { fmt.Println(initUsage) }
	flag.Parse()

	if len(flag.Args()) == 0 {
		fmt.Println(initUsage)
		return
	}

	if err := builder.initWorkspace(flag.Args()[0]); err != nil {
		log("ERROR", err.Error())
		builder.cleanInitBuilder()
		os.Exit(1)
	}

	log("SUCCESS", "You can know build & deploy application using 'tge-cli build' command (see help)")
}

var initUsage = `tge-cli init creates a TGE workspace.
	
Usage:
    tge-cli init package

Package argument can be of several forms:
    local   ex: my-app
    url     ex: github.com/me/my-app
	
In both cases, the last token will be used as worspace root.`
