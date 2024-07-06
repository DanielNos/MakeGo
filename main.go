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
	BUILD_FOLDER   = "bin"
	PACKAGE_FOLDER = ".makego_package"
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
var packageFormatCount int = 0
var packageIndex = 1
var configFile string = "make.toml"
var config Config

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

func fileName(platform string) string {
	splitPlatform := strings.Split(platform, "/")
	return config.Application.Name + "_" + config.Application.Version + "_" + splitPlatform[0] + "_" + splitPlatform[1]
}

func platformArchitecture(platformArchitecture string) (string, string) {
	split := strings.Split(platformArchitecture, "/")
	return split[0], split[1]
}

func countPackageFormats() {
	if config.Package.Apt {
		packageFormatCount++
	}
	if config.Package.Rpm {
		packageFormatCount++
	}
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
			"url = \"https://github.com/Username/app\"\n" +
			"license = \"\"\n" +

			"[build]\n" +
			"target = \".\"\n" +
			"flags = \"-ldflags=\\\"-w -s\\\"\"\n" +
			"platforms = [ \"linux/amd64\", \"windows/amd64\", \"darwin/arm64\" ]\n\n" +

			"[maintainer]\n" +
			"name = \"Name Surname\"\n" +
			"email = \"name.surname@email.com\"\n\n" +

			"[package]\n" +
			"apt = true\n",
	)
	if err != nil {
		logStepError(1, packageFormatCount, "Failed to write config: "+err.Error())
		return
	}
}

func checkRequirments() bool {
	// Check if OS is linux
	if runtime.GOOS != "linux" {
		logError(3, "Can't package on non-linux system.")
		return false
	}

	// Check if Ruby is installed
	cmd := exec.Command("ruby", "--version")
	_, err := cmd.CombinedOutput()
	if err != nil {
		logError(3, "Can't package RPM without Ruby installed.")
		return false
	}

	// Check if FPM is installed
	cmd = exec.Command("fpm", "--version")
	_, err = cmd.CombinedOutput()
	if err != nil {
		logError(3, "Can't package RPM without FPM installed.")
		return false
	}

	return true
}

func createFPMConfig() bool {
	file, err := os.Create(BUILD_FOLDER + "/.fpm")
	if err != nil {
		logStepError(1, packageFormatCount, "Failed to create config file: "+err.Error())
		return false
	}
	defer file.Close()

	flags := "-s dir\n" +
		"--name " + config.Application.Name + "\n" +
		"--version " + config.Application.Version + "\n" +
		"--description \"" + config.Application.Description + "\"\n" +
		"--url \"" + config.Application.Url + "\"\n" +
		"--maintainer \"" + config.Maintainer.Name + " <" + config.Maintainer.Email + ">\"\n"

	if strings.TrimSpace(config.Application.License) != "" {
		flags += "--license \"" + config.Application.License + "\"\n"
	}

	flags += fileName("linux/amd64") + "=/usr/bin/" + config.Application.Name + "\n"

	_, err = file.WriteString(flags)
	if err != nil {
		logStepError(1, packageFormatCount, "Failed to write config file: "+err.Error())
		return false
	}

	return true
}

func clean() {
	log(1, "Cleaning")
	os.RemoveAll(PACKAGE_FOLDER)
	os.RemoveAll(BUILD_FOLDER)
}

func buildBinary() {
	log(2, "Building binaries")
	os.Mkdir("bin", 0755)

	for i, target := range config.Build.Platforms {
		logStep(i+1, len(config.Build.Platforms), "Building target "+target)

		splitTarget := strings.Split(target, "/")
		outputPath := BUILD_FOLDER + "/" + fileName(target)

		if splitTarget[0] == "windows" {
			outputPath += ".exe"
		}

		cmd := exec.Command("go", "build", "-o", outputPath, config.Build.Target)
		cmd.Env = append(cmd.Environ(), "GOOS="+splitTarget[0], "GOARCH="+splitTarget[1])

		output, err := cmd.CombinedOutput()

		if err != nil {
			logStepError(i+1, len(config.Build.Platforms), string(output))
		}
	}
}

func packageApt() {
	logStep(packageIndex, packageFormatCount, "Packaging APT")
	packageIndex++

	// Check if dpkg-deb is installed
	cmd := exec.Command("dpkg-deb", "--version")
	_, err := cmd.CombinedOutput()
	if err != nil {
		logError(packageIndex-1, "Can't package APT without dpkg-deb installed.")
		return
	}

	// Find packageable binaries
	platforms := []string{}

	for _, platArch := range config.Build.Platforms {
		platform, architecture := platformArchitecture(platArch)

		if platform != "linux" {
			continue
		}

		if !validDebArch(architecture) {
			continue
		}

		platforms = append(platforms, platArch)
	}

	// Create packages
	for i, platArch := range platforms {
		logSubStep(i+1, len(platforms), "Packaging "+platArch+" deb")
		_, architecture := platformArchitecture(platArch)

		cmd = exec.Command("fpm", "-t", "deb", "--architecture", goArchToDebArch(architecture), fileName(platArch))
		cmd.Dir = BUILD_FOLDER
		output, err := cmd.CombinedOutput()
		if err != nil {
			logError(packageIndex-1, "Failed to package APT. "+string(output))
			return
		}
	}
}

func packageRpm() {
	logStep(packageIndex, packageFormatCount, "Packaging RPM")
	packageIndex++

	// Check if dpkg-deb is installed
	cmd := exec.Command("rpm", "--version")
	_, err := cmd.CombinedOutput()
	if err != nil {
		logError(packageIndex-1, "Can't package RPM without rpm installed.")
		return
	}

	cmd = exec.Command("fpm", "-t", "rpm")
	cmd.Dir = BUILD_FOLDER
	output, err := cmd.CombinedOutput()
	if err != nil {
		logError(packageIndex-1, "Failed to package RPM. "+string(output))
		return
	}
}

func createPackages() {
	log(3, "Packaging")

	meetsRequirements := checkRequirments()
	if !meetsRequirements {
		return
	}

	createFPMConfig()

	// Package
	if config.Package.Apt {
		packageApt()
	}
	if config.Package.Rpm {
		packageRpm()
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

	loadConfig()

	start := time.Now()
	info(start, "Building \""+config.Build.Target+"\"")

	make()

	success(fmt.Sprintf("Build complete in %s", time.Since(start)))
}
