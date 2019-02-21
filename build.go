package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path"
	"runtime"

	decentcopy "github.com/hugocarreira/go-decent-copy"
	"github.com/otiai10/copy"
)

func (builder *Builder) initBuilder(packagePath string) error {
	if !path.IsAbs(packagePath) {
		builder.packagePath = path.Join(builder.cwd, packagePath)
	} else {
		builder.packagePath = packagePath
	}

	if _, err := os.Stat(builder.packagePath); os.IsNotExist(err) {
		return fmt.Errorf("package path '%s' not found", builder.packagePath)
	}

	builder.programName = path.Base(builder.packagePath)
	builder.packageName = builder.programName

	if err := os.Chdir(builder.packagePath); err != nil {
		return err
	}

	if err := builder.installTGE(); err != nil {
		return err
	}

	builder.distPath = path.Join(builder.packagePath, distPath, builder.target)

	if !builder.devMode {
		if err := builder.cleanBuilBuilder(); err != nil {
			log("WARNING", fmt.Sprintf("failed to clean build: %s", err))
		}
	}

	if _, err := os.Stat(builder.distPath); os.IsNotExist(err) {
		if err = os.MkdirAll(builder.distPath, os.ModeDir|0622); err != nil {
			return err
		}
	}

	builder.assetsPath = path.Join(builder.packagePath, assetsPath)

	if _, err := os.Stat(builder.assetsPath); os.IsNotExist(err) {
		if err = os.MkdirAll(builder.assetsPath, os.ModeDir|0622); err != nil {
			return err
		}
	}

	return nil
}

func (builder *Builder) checkCopyResources() error {
	resourcesInPath := path.Join(builder.packagePath, builder.target)
	var err error
	if _, err = os.Stat(resourcesInPath); os.IsNotExist(err) {
		if err = os.MkdirAll(resourcesInPath, os.ModeDir|0722); err != nil {
			return err
		}
		if err = copy.Copy(path.Join(builder.tgeRootPath, tgeTemplatePath, builder.target), resourcesInPath); err != nil {
			return err
		}
		log("NOTICE", fmt.Sprintf("'%s' folder has been added to your project for customization (see README.md inside)", builder.target))
	}
	return nil
}

func (builder *Builder) installGoMobile() (string, error) {
	gomobilebin, err := exec.LookPath("gomobile")
	if err != nil {
		gomobilebin = path.Join(builder.goPath, "bin", "gomobile")
		if _, err = os.Stat(gomobilebin); os.IsNotExist(err) {
			log("NOTICE", "installing gomobile in your workspace")
			cmd := exec.Command("go", "get", "golang.org/x/mobile/cmd/gomobile")
			cmd.Env = append(os.Environ(),
				fmt.Sprintf("GOPATH=%s", builder.goPath),
			)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return "", fmt.Errorf("failed to install gomobile")
			}
		}
	}

	if builder.target == "android" {
		if _, err = os.Stat(path.Join(builder.goPath, "pkg", "gomobile", "ndk-toolchains")); os.IsNotExist(err) {
			androidHome := os.Getenv("ANDROID_HOME")
			if androidHome == "" {
				return "", fmt.Errorf("ANDROID_HOME is not set (must point to Android SDK)")
			}

			androidNDKPath := path.Join(androidHome, "ndk-bundle")

			if _, err = os.Stat(androidNDKPath); os.IsNotExist(err) {
				return "", fmt.Errorf("ANDROID NDK not found in %s, see https://developer.android.com/ndk/guides/", androidNDKPath)
			}

			log("NOTICE", "initializing gomobile for Android")
			cmd := exec.Command(gomobilebin, "init", "-ndk", androidNDKPath)
			cmd.Env = append(os.Environ(),
				fmt.Sprintf("GOPATH=%s", builder.goPath),
			)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return "", fmt.Errorf("failed to initialize gomobile")
			}
		}
	} else {
		log("NOTICE", "initializing gomobile for IOS")
		cmd := exec.Command(gomobilebin, "init")
		cmd.Env = append(os.Environ(),
			fmt.Sprintf("GOPATH=%s", builder.goPath),
		)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return "", fmt.Errorf("failed to initialize gomobile")
		}
	}
	return gomobilebin, nil
}

func (builder *Builder) buildAndroid(packagePath string) error {
	// Init
	builder.target = "android"

	if err := builder.initBuilder(packagePath); err != nil {
		return err
	}

	var gomobilebin string
	var err error
	if gomobilebin, err = builder.installGoMobile(); err != nil {
		return err
	}

	// Resources
	if err := builder.checkCopyResources(); err != nil {
		return fmt.Errorf("failed to copy resources files: %s", err)
	}

	if _, err := os.Stat(path.Join(builder.packagePath, "android", "AndroidManifest.xml")); os.IsNotExist(err) {
		if err = decentcopy.Copy(path.Join(builder.tgeRootPath, tgeTemplatePath, "android", "AndroidManifest.xml"), path.Join(builder.packagePath, "AndroidManifest.xml")); err != nil {
			return fmt.Errorf("WARNING", "failed to copy AndroidManifest.xml from TGE, using default gombile one: %s", err)
		}
	} else {
		if err = decentcopy.Copy(path.Join(builder.packagePath, builder.target, "AndroidManifest.xml"), path.Join(builder.packagePath, "AndroidManifest.xml")); err != nil {
			return fmt.Errorf("WARNING", "failed to copy AndroidManifest.xml, using default gombile one: %s", err)
		}
	}
	defer os.Remove(path.Join(builder.packagePath, "AndroidManifest.xml"))

	if builder.devMode {
		var cmd *exec.Cmd
		if builder.verbose {
			cmd = exec.Command(gomobilebin, "build", "-v", "-target=android", "-o", path.Join(builder.distPath, fmt.Sprintf("%s.apk", builder.programName)))
		} else {
			cmd = exec.Command(gomobilebin, "build", "-target=android", "-o", path.Join(builder.distPath, fmt.Sprintf("%s.apk", builder.programName)))
		}
		cmd.Env = append(os.Environ(),
			fmt.Sprintf("GOPATH=%s", builder.goPath),
		)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to build android application")
		}
	} else {
		for _, t := range []string{"arm", "386", "amd64", "arm64"} {
			var cmd *exec.Cmd
			if builder.verbose {
				cmd = exec.Command(gomobilebin, "build", "-v", fmt.Sprintf("-target=android/%s", t), "-o", path.Join(builder.distPath, fmt.Sprintf("%s-%s.apk", builder.programName, t)))
			} else {
				cmd = exec.Command(gomobilebin, "build", fmt.Sprintf("-target=android/%s", t), "-o", path.Join(builder.distPath, fmt.Sprintf("%s-%s.apk", builder.programName, t)))
			}
			cmd.Env = append(os.Environ(),
				fmt.Sprintf("GOPATH=%s", builder.goPath),
			)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to build android application (arch %s)", t)
			}
		}

	}

	return nil
}

func (builder *Builder) buildIOS(packagePath string, bundleID string) error {
	// Resources
	builder.target = "ios"

	if err := builder.initBuilder(packagePath); err != nil {
		return err
	}

	var gomobilebin string
	var err error
	if gomobilebin, err = builder.installGoMobile(); err != nil {
		return err
	}

	// Resources
	if err := builder.checkCopyResources(); err != nil {
		return fmt.Errorf("failed to copy resources files: %s", err)
	}

	// Build
	var cmd *exec.Cmd
	if builder.verbose {
		cmd = exec.Command(gomobilebin, "build", "-v", "-target=ios", fmt.Sprintf("-bundleid=%s", bundleID), "-o", path.Join(builder.distPath, fmt.Sprintf("%s.app", builder.programName)))
	} else {
		cmd = exec.Command(gomobilebin, "build", "-target=ios", fmt.Sprintf("-bundleid=%s", bundleID), "-o", path.Join(builder.distPath, fmt.Sprintf("%s.app", builder.programName)))
	}
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("GOPATH=%s", builder.goPath),
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to build IOS application")
	}

	return nil
}

func (builder *Builder) buildBrowser(packagePath string) error {
	// Init
	builder.target = "browser"

	if err := builder.initBuilder(packagePath); err != nil {
		return err
	}

	// Build
	var cmd *exec.Cmd
	if builder.verbose {
		cmd = exec.Command("go", "build", "-v", "-o", path.Join(builder.distPath, "main.wasm"))
	} else {
		cmd = exec.Command("go", "build", "-o", path.Join(builder.distPath, "main.wasm"))
	}
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("GOPATH=%s", builder.goPath),
		"GOOS=js",
		"GOARCH=wasm",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log("ERROR", "failed to build browser application")
		return fmt.Errorf("failed to build application")
	}
	// Resources
	if err := builder.checkCopyResources(); err != nil {
		return fmt.Errorf("failed to retrieve resources files from TGE: %s", err)
	}

	if err := copy.Copy(path.Join(builder.packagePath, builder.target), builder.distPath); err != nil {
		return fmt.Errorf("failed to copy resources files to dist: %s", err)
	}

	// Assets
	assetsOutPath := path.Join(builder.distPath, assetsPath)
	if _, err := os.Stat(assetsOutPath); os.IsNotExist(err) {
		if err := os.MkdirAll(assetsOutPath, os.ModeDir|0722); err != nil {
			return err
		}
		if err := copy.Copy(builder.assetsPath, assetsOutPath); err != nil {
			return err
		}
		log("NOTICE", fmt.Sprintf("Copying assets to dist: %s", assetsOutPath))
	} else if !builder.devMode {
		log("NOTICE", fmt.Sprintf("Copying assets to dist: %s", assetsOutPath))
		if err := copy.Copy(builder.assetsPath, assetsOutPath); err != nil {
			return err
		}
	} else {
		log("NOTICE", fmt.Sprintf("Skipping assets (DEV mode), found in dist: %s", assetsOutPath))
	}

	return nil
}

func (builder *Builder) buildDesktop(packagePath string) error {
	// Init
	switch runtime.GOOS {
	case "darwin":
		builder.target = "darwin"
	case "windows":
		builder.target = "windows"
	case "linux":
		builder.target = "linux"
	default:
		return fmt.Errorf("unsupported desktop target: '%s'", runtime.GOOS)
	}

	if err := builder.initBuilder(packagePath); err != nil {
		return err
	}

	binaryFile := builder.programName
	if builder.target == "windows" {
		binaryFile = fmt.Sprintf("%s.exe", binaryFile)
	}

	// Resources
	if err := builder.checkCopyResources(); err != nil {
		return fmt.Errorf("failed to retrieve resources files from TGE : %s", err)
	}

	// Build & packaging
	var cmd *exec.Cmd
	var assetsOutPath string
	switch builder.target {
	case "darwin":
		// Build
		if builder.verbose {
			cmd = exec.Command("go", "build", "-v", "-o", path.Join(builder.distPath, binaryFile))
		} else {
			cmd = exec.Command("go", "build", "-o", path.Join(builder.distPath, binaryFile))
		}
		cmd.Env = append(os.Environ(),
			fmt.Sprintf("GOPATH=%s", builder.goPath),
		)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to build application")
		}

		// Packaging
		appifybin, err := exec.LookPath("appify")
		if err != nil {
			appifybin = path.Join(builder.goPath, "bin", "appify")
			if _, err = os.Stat(appifybin); os.IsNotExist(err) {
				log("NOTICE", "installing appify in your workspace")
				cmd = exec.Command("go", "get", "github.com/machinebox/appify")
				cmd.Env = append(os.Environ(),
					fmt.Sprintf("GOPATH=%s", builder.goPath),
				)
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				if err := cmd.Run(); err != nil {
					appifybin = ""
					log("WARNING", "failed to install appify, unable to package MacOS application")
				}
			}
		}

		if appifybin != "" {
			os.Chdir(builder.distPath)
			cmd := exec.Command(appifybin, "-name", builder.programName, "-icon",
				path.Join(builder.packagePath, builder.target, "icon.png"), path.Join(builder.distPath, builder.programName))
			cmd.Env = append(os.Environ(),
				fmt.Sprintf("GOPATH=%s", builder.goPath),
			)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				log("WARNING", "failed to package MacOS application")
			}
		}

		os.RemoveAll(path.Join(builder.distPath, binaryFile))

		// Assets
		assetsOutPath = path.Join(builder.distPath, fmt.Sprintf("%s.app", builder.programName), "Contents", "Resources")
		log("NOTICE", fmt.Sprintf("Copying assets in dist: %s", assetsOutPath))
		if err := copy.Copy(builder.assetsPath, assetsOutPath); err != nil {
			return err
		}

	case "windows":
		// Packaging
		goversioninfobin, err := exec.LookPath("goversioninfo.exe")
		if err != nil {
			goversioninfobin = path.Join(builder.goPath, "bin", "goversioninfo.exe")
			if _, err = os.Stat(goversioninfobin); os.IsNotExist(err) {
				log("NOTICE", "installing goversioninfo in your workspace")
				cmd = exec.Command("go", "get", "github.com/josephspurrier/goversioninfo/cmd/goversioninfo")
				cmd.Env = append(os.Environ(),
					fmt.Sprintf("GOPATH=%s", builder.goPath),
				)
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				if err := cmd.Run(); err != nil {
					goversioninfobin = ""
					log("WARNING", "failed to install goversioninfo, unable to package Windows application")
				}
			}
		}

		if goversioninfobin != "" {
			if err := decentcopy.Copy(path.Join(builder.packagePath, builder.target, "versioninfo.json"), path.Join(builder.packagePath, "versioninfo.json")); err != nil {
				log("WARNING", "failed to prepare package for Windows application")
			} else {
				defer os.Remove(path.Join(builder.packagePath, "resource_windows_386.syso"))
				defer os.Remove(path.Join(builder.packagePath, "resource_windows_amd64.syso"))
			}
			defer os.Remove(path.Join(builder.packagePath, "versioninfo.json"))

			cmd := exec.Command(goversioninfobin, "-platform-specific=true", "-manifest", path.Join(builder.packagePath, builder.target, "main.exe.manifest"), "-icon",
				path.Join(builder.packagePath, builder.target, "icon.ico"))
			cmd.Env = append(os.Environ(),
				fmt.Sprintf("GOPATH=%s", builder.goPath),
			)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				log("WARNING", "failed to prepare package for Windows application")
			}
		}

		// Build
		if builder.verbose {
			cmd = exec.Command("go", "build", "-ldflags", "-H=windowsgui", "-v", "-o", path.Join(builder.distPath, binaryFile))
		} else {
			cmd = exec.Command("go", "build", "-ldflags", "-H=windowsgui", "-o", path.Join(builder.distPath, binaryFile))
		}
		cmd.Env = append(os.Environ(),
			fmt.Sprintf("GOPATH=%s", builder.goPath),
		)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to build application")
		}

		// Assets
		assetsOutPath = path.Join(builder.distPath, assetsPath)
		if _, err := os.Stat(assetsOutPath); os.IsNotExist(err) {
			if err := os.MkdirAll(assetsOutPath, os.ModeDir|0722); err != nil {
				return err
			}
			if err := copy.Copy(builder.assetsPath, assetsOutPath); err != nil {
				return err
			}
			log("NOTICE", fmt.Sprintf("Copying assets in dist: %s", assetsOutPath))
		} else if !builder.devMode {
			log("NOTICE", fmt.Sprintf("Copying assets in dist: %s", assetsOutPath))
			if err := copy.Copy(builder.assetsPath, assetsOutPath); err != nil {
				return err
			}
		} else {
			log("NOTICE", fmt.Sprintf("Skipping assets (DEV mode), found in dist: %s", assetsOutPath))
		}
	}

	return nil
}

func (builder *Builder) cleanBuilBuilder() error {
	if builder.distPath != "" {
		return os.RemoveAll(builder.distPath)
	}
	return nil
}

func doBuild(builder Builder) {
	targetFlag := flag.String("target", "desktop", "build target : desktop, android, ios, browser")
	verboseFlag := flag.Bool("v", false, "verbose ouput for debugging")
	devModeFlag := flag.Bool("dev", false, "Dev mode, skip clean, assets copy & arch split (faster)")
	bundleIDFlag := flag.String("bundleid", "", "IOS only, bundleId to use for app")
	os.Args = os.Args[1:]
	flag.Usage = func() { fmt.Println(buildUsage) }
	flag.Parse()

	if len(flag.Args()) == 0 {
		fmt.Println(buildUsage)
		return
	}

	builder.devMode = *devModeFlag
	builder.verbose = *verboseFlag
	switch *targetFlag {
	case "desktop":
		if err := builder.buildDesktop(flag.Args()[0]); err != nil {
			log("ERROR", err.Error())
			builder.cleanBuilBuilder()
			os.Exit(1)
		}
	case "browser":
		if err := builder.buildBrowser(flag.Args()[0]); err != nil {
			log("ERROR", err.Error())
			builder.cleanBuilBuilder()
			os.Exit(1)
		}
	case "android":
		if err := builder.buildAndroid(flag.Args()[0]); err != nil {
			log("ERROR", err.Error())
			builder.cleanBuilBuilder()
			os.Exit(1)
		}
	case "ios":
		if *bundleIDFlag == "" {
			log("ERROR", "missing bundleId for IOS (set with -bundleid)")
			os.Exit(1)
		}
		if err := builder.buildIOS(flag.Args()[0], *bundleIDFlag); err != nil {
			log("ERROR", err.Error())
			builder.cleanBuilBuilder()
			os.Exit(1)
		}
	default:
		log("ERROR", fmt.Sprintf("unsupported target '%s'", *targetFlag))
		os.Exit(1)
	}

	log("SUCCESS", fmt.Sprintf("Application is available in %s", builder.distPath))
}

var buildUsage = `tge-cli build build and deploys TGE applications.
	
Usage:
	tge-cli build [-target TARGET] [-dev] [-v] [-bundleid ID] packagePath

The package path must point to a valid TGE application, the generated
application will be stored in the dist/$TARGET folder.

-target defines the application target:
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

-dev flag allows to generate application faster by omitting assets copy,
on Android the resulting APK will support all architectures.

-v verbose output for debugging purpose

-bundleid is mandatory for IOS build and can be obtained from Apple Developer.`
