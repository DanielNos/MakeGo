package main

import (
	"fmt"
	"os"
	"os/exec"
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
}

func countPackageFormats() {
	if config.Package.Apt {
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
		outputPath := BUILD_FOLDER + "/" + config.Application.Name + "_" + splitTarget[0] + "_" + splitTarget[1]

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
	logStep(1, packageFormatCount, "Packaging APT")

	packageName := config.Application.Name + "_" + config.Application.Version + "-amd64"
	packagePath := PACKAGE_FOLDER + "/" + packageName

	// Create directories
	os.MkdirAll(packagePath+"/usr/bin", 0755)
	os.Mkdir(packagePath+"/DEBIAN", 0755)

	// Create control file
	file, err := os.Create(packagePath + "/DEBIAN/control")
	if err != nil {
		logStepError(1, packageFormatCount, "Failed to create control: "+err.Error())
		return
	}
	defer file.Close()

	_, err = file.WriteString(
		"Package: " + config.Application.Name + "\n" +
			"Version: " + config.Application.Version + "\n" +
			"Architecture: amd64\n" +
			"Maintainer: " + config.Maintainer.Name + " <" + config.Maintainer.Email + ">\n" +
			"Description: " + config.Application.Description + "\n",
	)
	if err != nil {
		logStepError(1, packageFormatCount, "Failed to write control: "+err.Error())
		return
	}

	// Copy binary
	binaryDestination := packagePath + "/usr/bin/" + config.Application.Name
	err = copyFile(BUILD_FOLDER+"/"+config.Application.Name+"_linux_amd64", binaryDestination)
	if err != nil {
		logStepError(1, packageFormatCount, "Failed to copy binary: "+err.Error())
		return
	}

	err = os.Chmod(binaryDestination, 0755)
	if err != nil {
		logStepError(1, packageFormatCount, "Failed to change permissions: "+err.Error())
		return
	}

	// Create package
	cmd := exec.Command("dpkg-deb", "--build", "--root-owner-group", packagePath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		logStepError(1, packageFormatCount, "Failed to generate package: "+string(output))
		return
	}

	err = os.Rename(packagePath+".deb", BUILD_FOLDER+"/"+packageName+".deb")
	if err != nil {
		logStepError(1, packageFormatCount, "Failed to move package: "+err.Error())
		return
	}
}

func createPackages() {
	log(3, "Packaging")

	os.Mkdir(PACKAGE_FOLDER, 0755)

	if config.Package.Apt {
		packageApt()
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
