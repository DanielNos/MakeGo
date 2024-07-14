package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
)

const VERSION = "0.1.0"

const (
	BIN_DIR     = "bin"
	PKG_DIR     = "pkg"
	DEB_PKG_DIR = PKG_DIR + "/.Deb"
	RPM_PKG_DIR = PKG_DIR + "/.rpm"
)

type Action uint8

const (
	A_None Action = iota
	A_Generate
	A_Clean
	A_Binary
	A_Package
	A_All
)

var action Action
var configFile string
var config Config

var packageFormatCount int = 0
var packageIndex = 1

var generateTarget string

var stringToAction = map[string]Action{
	"new":     A_Generate,
	"clean":   A_Clean,
	"cln":     A_Clean,
	"binary":  A_Binary,
	"bin":     A_Binary,
	"package": A_Package,
	"pkg":     A_Package,
	"all":     A_All,
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
	if config.Deb.Package {
		packageFormatCount++
	}
	if config.RPM.Package {
		packageFormatCount++
	}
}

func isBuildArch(arch string) bool {
	for _, platArch := range config.Build.Platforms {
		if strings.Split(platArch, "/")[1] == arch {
			return true
		}
	}
	return false
}

func printHelp() {
	fmt.Println("MakeGo\n")
	fmt.Println("Usage: makego [action] [config]\n")
	fmt.Println("Actions:")
	fmt.Println("      help           Shows help.")
	fmt.Println("      new [template] Creates a config template. Templates: default (or none), all.")
	fmt.Println("      cln/clean      Removes all build and package files.")
	fmt.Println("      bin/binary     Builds binaries.")
	fmt.Println("      pkg/package    Builds binaries and packages them.")
	fmt.Println("      all (or none)  Does cln -> bin -> pkg.\n")
	fmt.Println("Flags:")
	fmt.Println("      -h --help     Show help.")
	fmt.Println("      -v --version  Show version.")
	fmt.Println("      -t --time     Print time stamps.")
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
			printHelp()
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
					if maybeAction == A_Generate {
						generateTarget = "*"
					}

					action = maybeAction
				}
			}
		}
	}

	if action == A_None {
		action = A_All
	}
	if configFile == "" {
		configFile = "make.toml"
	}
	if generateTarget == "*" {
		generateTarget = "normal"
	}
}

func loadConfig() {
	_, err := toml.DecodeFile(configFile, &config)

	if err != nil {
		fatal(fmt.Sprintf("Failed to load make config \"%s\": %s", configFile, strings.Split(err.Error(), ":")[1][1:]))
	}

	countPackageFormats()
}

func writeDefaultConfig() {
	configText := CONFIG_DEFAULT
	if generateTarget == "all" {
		configText = CONFIG_ALL
	} else {
		generateTarget = "default"
	}

	info(time.Now(), "Generating config template "+generateTarget+" to "+configFile+".")

	file, err := os.Create(configFile)
	if err != nil {
		fatal("Failed to create config: " + err.Error())
		return
	}
	defer file.Close()

	_, err = file.WriteString(configText)
	if err != nil {
		stepError("Failed to write config: "+err.Error(), 1, packageFormatCount, 0)
		return
	}
}

func checkRequirements() bool {
	if runtime.GOOS != "linux" {
		stepError("Can't package on non-linux system.", 3, 3, 0)
		return false
	}
	return true
}

func clean() {
	step("Cleaning", 1, 3, 0, false)
	os.RemoveAll(PKG_DIR)
	os.RemoveAll(BIN_DIR)
}

func buildBinaries() {
	step("Building binaries", 2, 3, 0, false)

	cmd := exec.Command("go", "get")
	output, err := cmd.CombinedOutput()
	if err != nil {
		stepError("Failed to run get dependecies. "+string(output), 1, 3, 0)
	}

	os.Mkdir("bin", 0755)

	for i, target := range config.Build.Platforms {
		step("Building target "+target, i+1, len(config.Build.Platforms), 1, true)

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
	step("Packaging", 3, 3, 0, false)

	meetsRequirements := checkRequirements()
	if !meetsRequirements {
		return
	}

	os.MkdirAll(PKG_DIR, 0755)

	// Package
	if config.Deb.Package {
		packageDeb()
	}

	if config.RPM.Package {
		packageRPM()
	}
}

func make() {
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

	if action == A_Generate {
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

	make()

	success(fmt.Sprintf("Build complete in %s", time.Since(start)))
}
