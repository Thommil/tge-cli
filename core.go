package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const tgeVersion = "master"
const tgeLocalGoPath = ".tge"
const tgeTemplatePath = "template"

var tgePackageName = "github.com/thommil/tge"

const distPath = "dist"
const assetsPath = "assets"

type Builder struct {
	//all
	cwd         string
	packageName string
	packagePath string
	goPath      string
	tgeRootPath string
	verbose     bool

	//build
	target      string
	devMode     bool
	assetsPath  string
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
	builder.goPath = os.Getenv("GOPATH")
	if builder.goPath == "" {
		builder.goPath = filepath.Join(builder.packagePath, tgeLocalGoPath)
	}

	cmd := exec.Command("go", "list", "-e", "-f", "{{.Dir}}", tgePackageName)
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("GOPATH=%s", builder.goPath),
	)
	if output, err := cmd.Output(); err != nil {
		return fmt.Errorf("failed to analyze GOPATH %s: %s", builder.goPath, err)
	} else {
		builder.tgeRootPath = strings.TrimSpace(string(output))
	}

	if builder.tgeRootPath == "" {
		if _, err := os.Stat(filepath.Join(builder.packagePath, "go.mod")); os.IsNotExist(err) {
			log("NOTICE", fmt.Sprintf("Initializing '%s' module", builder.packageName))
			cmd := exec.Command("go", "mod", "init", builder.packageName)
			cmd.Env = append(os.Environ(),
				fmt.Sprintf("GOPATH=%s", builder.goPath),
			)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to initialze workspace")
			}
		}

		log("NOTICE", fmt.Sprintf("Installing TGE in %s", builder.goPath))
		log("NOTICE", fmt.Sprintf("Using GOPATH %s (set it for DEV)", builder.goPath))
		cmd := exec.Command("go", "get", "-u", tgePackageName)
		cmd.Env = append(os.Environ(),
			fmt.Sprintf("GOPATH=%s", builder.goPath),
		)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to install TGE")
		}

		cmd = exec.Command("go", "list", "-e", "-f", "{{.Dir}}", tgePackageName)
		cmd.Env = append(os.Environ(),
			fmt.Sprintf("GOPATH=%s", builder.goPath),
		)
		if output, err := cmd.Output(); err != nil {
			return fmt.Errorf("failed to analyze GOPATH %s: %s", builder.goPath, err)
		} else {
			builder.tgeRootPath = strings.TrimSpace(string(output))
		}

		if builder.tgeRootPath == "" {
			return fmt.Errorf("failed to install TGE, try manually using 'go get %s'", tgePackageName)
		}
	}

	if err := filepath.Walk(builder.tgeRootPath, func(p string, info os.FileInfo, err error) error {
		if err = os.Chmod(p, info.Mode()|os.FileMode(0222)); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return fmt.Errorf("failed to set read permission on %s: %s", builder.tgeRootPath, err)
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
	if minor < 12 {
		err = fmt.Errorf("Go 1.12 or newer is required")
		log("ERROR", err.Error())
		return err
	}
	return nil
}
