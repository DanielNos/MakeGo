package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const VERSION = "0.1.0"

const HELP = `MakeGo

Usage: makego [action] [config]

Actions:
   help           Shows help.
   new [template] Creates a config template. Templates: default (or none), all, empty.
   cln/clean      Removes all build and package files.
   bin/binary     Builds binaries.
   pkg/package    Builds binaries and packages them.
   all (or none)  Does cln -> bin -> pkg.

Flags:
    -h --help     Show help.
    -v --version  Show version.
    -t --time     Print time stamps.
`

const (
	BUILD_DIR   = "build"
	UTILITY_DIR = ".makego"

	BIN_DIR = BUILD_DIR + "/bin"
	PKG_DIR = BUILD_DIR + "/pkg"

	SRC_PKG_DIR      = PKG_DIR + "/.src"
	DEB_PKG_DIR      = PKG_DIR + "/.deb"
	RPM_PKG_DIR      = PKG_DIR + "/.rpm"
	PKG_PKG_DIR      = PKG_DIR + "/.pkg"
	APPIMAGE_PKG_DIR = PKG_DIR + "/.appimage"
)

type Action uint8

const (
	A_None Action = iota
	A_New
	A_Clean
	A_Binary
	A_Package
)

var action Action
var configFile string
var config Config

var packageFormatCount int = 0
var packageIndex = 1

var generateTarget string

var stringToAction = map[string]Action{
	"new":     A_New,
	"clean":   A_Clean,
	"cln":     A_Clean,
	"binary":  A_Binary,
	"bin":     A_Binary,
	"package": A_Package,
	"pkg":     A_Package,
	"all":     A_Package,
}

func b2i(b bool) int {
	if b {
		return 1
	}
	return 0
}

func isInstalled(packageName string) bool {
	cmd := exec.Command(packageName, "--version")
	_, err := cmd.CombinedOutput()
	return err == nil
}

func fileName(platArch string) string {
	splitPlatform := strings.Split(platArch, "/")
	return config.Application.Name + "_" + config.Application.Version + "_" + splitPlatform[0] + "_" + splitPlatform[1]
}

func splitPlatArch(platformArchitecture string) (string, string) {
	split := strings.Split(platformArchitecture, "/")
	return split[0], split[1]
}

func countPackageFormats() {
	packageFormatCount = b2i(config.Deb.Package) + b2i(config.RPM.Package) + b2i(config.Pkg.Package) + b2i(config.AppImage.Package)
}

func isBuildArch(arch string) bool {
	for _, platArch := range config.Build.Platforms {
		if strings.Split(platArch, "/")[1] == arch {
			return true
		}
	}
	return false
}

func compressSource() error {
	sourcePath := SRC_PKG_DIR + "/" + config.Application.Name + "-" + config.Application.Version
	sourcePath, _ = filepath.Abs(sourcePath)
	os.MkdirAll(sourcePath, 0755)

	// Copy source files to temp
	cmd := exec.Command("rsync", "-a",
		".",
		sourcePath,
		"--exclude", "bin", "--exclude", "pkg", "--exclude", ".git", "--exclude", ".vscode", "--exclude", "LICENSE",
	)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return errors.New("Failed to copy source: " + string(output))
	}

	// Compress source files
	cmd = exec.Command("tar",
		"-czf", sourcePath+".tar.gz",
		config.Application.Name+"-"+config.Application.Version,
	)
	cmd.Dir = SRC_PKG_DIR
	output, err = cmd.CombinedOutput()

	if err != nil {
		return errors.New("Failed to compress source: " + string(output))
	}

	return nil
}

func parseArguments() {
	action = A_None
	configFile = ""

	for i, arg := range os.Args[1:] {
		if generateTarget == "*" {
			if strings.HasSuffix(arg, ".toml") {
				if configFile != "" {
					fatal(fmt.Sprintf("argument %d: more than 1 config file specified.", i+1))
				}
				configFile = arg
			} else {
				generateTarget = arg
			}
			continue
		}

		switch arg {
		case "-h", "--help", "help":
			fmt.Print(HELP)
			os.Exit(0)

		case "-v", "--version":
			fmt.Println("MakeGo " + VERSION)
			os.Exit(0)

		case "-t", "--time":
			logTimeStamps = true

		default:
			if strings.HasSuffix(arg, ".toml") {
				if configFile != "" {
					fatal(fmt.Sprintf("argument %d: more than 1 config file specified.", i+1))
				}
				configFile = arg
			} else {
				maybeAction := stringToAction[arg]
				if maybeAction == A_None {
					configFile = arg
				} else {
					if maybeAction == A_New {
						generateTarget = "*"
					}

					action = maybeAction
				}
			}
		}
	}

	if action == A_None {
		action = A_Package
	}
	if configFile == "" {
		configFile = "make.toml"
	}
	if generateTarget == "*" {
		generateTarget = "normal"
	}
}

func checkRequirements() bool {
	if runtime.GOOS != "linux" {
		stepError("Can't package on a non-linux operating system.", 3, int(action)-1, 0)
		return false
	}
	return true
}

func clean() {
	step("Cleaning", 1, int(action)-1, 0, false)
	os.RemoveAll(PKG_DIR)
	os.RemoveAll(BIN_DIR)
	os.RemoveAll(BUILD_DIR)
}

func buildBinaries() {
	step("Building binaries", 2, int(action)-1, 0, false)

	cmd := exec.Command("go", "get")
	output, err := cmd.CombinedOutput()
	if err != nil {
		stepError("Failed to run get dependencies. "+string(output), 1, int(action)-1, 0)
	}

	os.Mkdir(BIN_DIR, 0755)

	for i, target := range config.Build.Platforms {
		step("Building platform "+target, i+1, len(config.Build.Platforms), 1, true)

		splitTarget := strings.Split(target, "/")
		outputPath := BIN_DIR + "/" + fileName(target)

		if splitTarget[0] == "windows" {
			outputPath += ".exe"
		}

		cmd := exec.Command("go", "build", "-o", outputPath, config.Build.Target)
		cmd.Env = append(cmd.Environ(), "GOOS="+splitTarget[0], "GOARCH="+splitTarget[1])

		output, err := cmd.CombinedOutput()

		if err != nil {
			stepError(string(output), i+1, len(config.Build.Platforms), 1)
		}
	}
}

func createPackages() {
	step("Packaging", 3, int(action)-1, 0, false)

	// Check requirements
	meetsRequirements := checkRequirements()
	if !meetsRequirements {
		return
	}

	// Compress source
	if action >= A_Package && (config.Deb.Package || config.RPM.Package || config.Pkg.Package) {
		compressSource()
	}

	// Package
	os.MkdirAll(PKG_DIR, 0755)

	if config.Deb.Package {
		packageDeb()
	}
	if config.RPM.Package {
		packageRPM()
	}
	if config.Pkg.Package {
		packagePkg()
	}
	if config.AppImage.Package {
		packageAppImage()
	}
}

func build() {
	clean()

	if action >= A_Binary {
		buildBinaries()
	}

	if action >= A_Package {
		createPackages()
	}
}

func main() {
	parseArguments()

	if action == A_New {
		writeDefaultConfig()
		return
	}

	cmd := exec.Command("go", "help")
	_, err := cmd.CombinedOutput()
	if err != nil {
		fatal("Can't build without go installed.")
	}

	loadConfig()

	start := time.Now()
	info(start, "Building \""+config.Build.Target+"\"")

	build()

	success(fmt.Sprintf("Build complete in %s", time.Since(start)))
}
