package main

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
)

const tgeVersion = "master"
const tgeLocalGoPath = ".tge"
const tgeTemplatePath = "template"

var tgePackageName = fmt.Sprintf("github.com/thommil/tge@%s", tgeVersion)
var tgeMarker = fmt.Sprintf("tge-%s.marker", tgeVersion)

const distPath = "dist"

type Builder struct {
	//all
	cwd         string
	packageName string
	packagePath string
	goPath      string
	tgeRootPath string
	verbose     bool

	//init

	//build
	target      string
	devMode     bool
	distPath    string
	programName string
}

func createBuilder() Builder {
	if err := checkGoVersion(); err != nil {
		panic(err)
	}

	builder := Builder{}
	builder.cwd, _ = os.Getwd()

	return builder
}

// Builder common
func (builder *Builder) installTGE() error {
	builder.goPath = path.Join(builder.packagePath, tgeLocalGoPath)

	if _, err := os.Stat(builder.goPath); err == nil {
		if err := filepath.Walk(builder.goPath, func(p string, info os.FileInfo, err error) error {
			if !info.IsDir() && info.Name() == tgeMarker {
				builder.tgeRootPath = path.Dir(p)
			}
			return nil
		}); err != nil {
			return fmt.Errorf("failed to analyze GOPATH %s: %s", builder.goPath, err)
		}
	}

	if builder.tgeRootPath == "" {
		if _, err := os.Stat(path.Join(builder.packagePath, "go.mod")); os.IsNotExist(err) {
			log("NOTICE", fmt.Sprintf("Initializing '%s' module", builder.packageName))
			cmd := exec.Command("go", "mod", "init", builder.packageName)
			cmd.Env = append(os.Environ(),
				"GO111MODULE=on",
				fmt.Sprintf("GOPATH=%s", builder.goPath),
			)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to initialze workspace")
			}
		}

		log("NOTICE", fmt.Sprintf("Installing TGE in %s", builder.packagePath))
		log("NOTICE", fmt.Sprintf("Using GOPATH %s (set it for DEV)", builder.goPath))
		cmd := exec.Command("go", "get", tgePackageName)
		cmd.Env = append(os.Environ(),
			"GO111MODULE=on",
			fmt.Sprintf("GOPATH=%s", builder.goPath),
		)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to install TGE")
		}

		if err := filepath.Walk(builder.goPath, func(p string, info os.FileInfo, err error) error {
			if !info.IsDir() && info.Name() == tgeMarker {
				builder.tgeRootPath = path.Dir(p)
			}
			return nil
		}); err != nil {
			return fmt.Errorf("failed to analyze GOPATH %s: %s", builder.goPath, err)
		}

		if builder.tgeRootPath == "" {
			return fmt.Errorf("failed to install TGE, try manually using 'go get %s'", tgePackageName)
		}

		if err := filepath.Walk(builder.tgeRootPath, func(p string, info os.FileInfo, err error) error {
			if err = os.Chmod(p, info.Mode()|os.FileMode(0200)); err != nil {
				return err
			}
			return nil
		}); err != nil {
			return fmt.Errorf("failed to set read permission on %s: %s", builder.tgeRootPath, err)
		}
	}

	return nil
}

// LOGS
func log(state string, msg string) {
	if state == "SUCCESS" {
		fmt.Printf("tge: %s\n%s\n", state, msg)
	} else {
		fmt.Printf("tge: %s %s\n", state, msg)
	}
}

// Helpers
func checkGoVersion() error {
	gobin, err := exec.LookPath("go")
	if err != nil {
		err = fmt.Errorf("go not found")
		log("ERROR", err.Error())
		return err
	}
	goVersionOut, err := exec.Command(gobin, "version").CombinedOutput()
	if err != nil {
		err = fmt.Errorf("'go version' failed: %v, %s", err, goVersionOut)
		log("ERROR", err.Error())
		return err
	}
	var minor int
	if _, err := fmt.Sscanf(string(goVersionOut), "go version go1.%d", &minor); err != nil {
		// Ignore unknown versions; it's probably a devel version.
		return nil
	}
	if minor < 11 {
		err = fmt.Errorf("Go 1.11 or newer is required")
		log("ERROR", err.Error())
		return err
	}
	return nil
}
