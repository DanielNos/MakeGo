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

const (
	BIN_DIR     = "bin"
	PKG_DIR     = "pkg"
	DEB_PKG_DIR = PKG_DIR + "/.deb"
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

var action Action = A_All
var configFile string = "make.toml"
var config Config

var packageFormatCount int = 0
var packageIndex = 1

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
	if config.DEB.Package {
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

func selectActionAndTarget(arguments []string) {
	if len(arguments) == 0 {
		return
	}

	actionString := arguments[0]

	if len(arguments) > 1 {
		configFile = arguments[0]
		actionString = arguments[1]
	}

	action = stringToAction[actionString]

	if action == A_None {
		action = A_All
		if strings.HasSuffix(actionString, ".toml") {
			configFile = actionString
			return
		}

		fatal("Unknown action " + actionString + ".")
	}
}

func loadConfig() {
	_, err := toml.DecodeFile(configFile, &config)

	if err != nil {
		fatal(fmt.Sprintf("Failed to load make config \"%s\": %s", configFile, strings.Split(err.Error(), ":")[1][1:]))
	}

	countPackageFormats()
}

func generateDefault() {
	file, err := os.Create(configFile)
	if err != nil {
		fatal("Failed to create config: " + err.Error())
		return
	}
	defer file.Close()

	_, err = file.WriteString(
		"[application]\n" +
			"name = \"app\"\n" +
			"version = \"1.0.0\"\n" +
			"description = \"My cool application.\"\n\n" +
			"long_description = \"My cool application.\"\n" +
			"url = \"https://github.com/Username/app\"\n" +
			"license = \"\"\n\n" +

			"[maintainer]\n" +
			"name = \"Name Surname\"\n" +
			"email = \"name.surname@email.com\"\n\n" +

			"[build]\n" +
			"target = \".\"\n" +
			"flags = \"-ldflags=\\\"-w -s\\\"\"\n" +
			"platforms = [ \"linux/amd64\", \"windows/amd64\", \"darwin/arm64\" ]\n\n" +

			"[DEB]\n" +
			"package = false\n" +
			"architectures = [ amd64 ]\n\n" +

			"[RPM]\n" +
			"package = false\n" +
			"build_src = true\n" +
			"architectures = [ amd64 ]\n",
	)
	if err != nil {
		stepError("Failed to write config: "+err.Error(), 1, packageFormatCount, 0)
		return
	}
}

func checkRequirments() bool {
	if runtime.GOOS != "linux" {
		stepError("Can't package on non-linux system.", 3, 3, 0)
		return false
	}
	return true
}

func clean() {
	step("Cleaning", 1, 3, 0)
	os.RemoveAll(PKG_DIR)
	os.RemoveAll(BIN_DIR)
}

func buildBinary() {
	step("Building binaries", 2, 3, 0)

	cmd := exec.Command("go", "get")
	output, err := cmd.CombinedOutput()
	if err != nil {
		stepError("Failed to run get dependecies. "+string(output), 1, 3, 0)
	}

	os.Mkdir("bin", 0755)

	for i, target := range config.Build.Platforms {
		step("Building target "+target, i+1, len(config.Build.Platforms), 1)

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
	step("Packaging", 3, 3, 0)

	meetsRequirements := checkRequirments()
	if !meetsRequirements {
		return
	}

	os.MkdirAll(PKG_DIR, 0755)

	// Package
	if config.DEB.Package {
		packageDEB()
	}

	if config.RPM.Package {
		packageRPM()
	}
}

func make() {
	clean()

	if action >= A_Binary {
		buildBinary()
	}

	if action >= A_Package {
		createPackages()
	}
}

func main() {
	selectActionAndTarget(os.Args[1:])

	if action == A_Generate {
		generateDefault()
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
