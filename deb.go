package main

import (
	"errors"
	"os"
	"os/exec"
)

func checkDebRequirements() bool {
	if !isInstalled("dpkg-deb") {
		stepError("Can't package deb without dpkg-Deb installed.", packageIndex-1, packageFormatCount, 1)
		return false
	}
	return true
}

func writeControlFile(arch string) {
	file, err := os.Create(DEB_PKG_DIR + "/" + config.Application.Name + "-" + config.Application.Version + "/DEBIAN/control")
	if err != nil {
		fatal("Failed to create control file: " + err.Error())
		return
	}
	defer file.Close()

	writeLine(file, "Package: "+config.Application.Name)
	writeLine(file, "Version: "+config.Application.Version)
	writeLine(file, "Architecture: "+goArchToPackageArch(arch))
	writeLine(file, "Maintainer: "+config.Maintainer.Name+" <"+config.Maintainer.Email+">")
	writeLine(file, "Description: "+config.Application.Description)
	writeLine(file, "Section: custom")
	writeLine(file, "Priority: optional")
}

func makeDebPackage(arch string) error {
	// Create control file
	writeControlFile(arch)

	// Remove old binary
	appName := config.Application.Name + "-" + config.Application.Version
	binDirectory := DEB_PKG_DIR + "/" + appName + "/usr/bin"
	cmd := exec.Command("rm", "-rf", binDirectory)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return errors.New("Failed to remove old binary: " + string(output))
	}
	os.Mkdir(binDirectory, 0755)

	// Copy binary
	cmd = exec.Command("cp",
		BIN_DIR+"/"+fileName("linux/"+arch),
		DEB_PKG_DIR+"/"+config.Application.Name+"-"+config.Application.Version+"/usr/bin/",
	)
	output, err = cmd.CombinedOutput()

	if err != nil {
		return errors.New("Failed to copy binary: " + string(output))
	}

	// Package
	cmd = exec.Command("dpkg-deb", "--build", DEB_PKG_DIR+"/"+appName)
	output, err = cmd.CombinedOutput()

	if err != nil {
		return errors.New("Failed to package binary: " + string(output))
	}

	// Rename package
	err = os.Rename(DEB_PKG_DIR+"/"+appName+".deb", "./"+PKG_DIR+"/"+appName+"-"+arch+".deb")
	if err != nil {
		return errors.New("Failed to rename package: " + err.Error())
	}

	return nil
}

func packageDeb() {
	step("Packaging deb", packageIndex, packageFormatCount, 1, false)
	packageIndex++

	// Check requirements
	if !checkDebRequirements() {
		return
	}

	// Create packagigng directories
	packagingDir := DEB_PKG_DIR + "/" + config.Application.Name + "-" + config.Application.Version
	makeDirs([]string{packagingDir + "/DEBIAN", packagingDir + "/usr/bin"}, 0755)

	// Create packages
	for i, arch := range config.Deb.Architectures {
		step("Packaging "+arch, i+1, len(config.Deb.Architectures), 2, true)

		if !isBuildArch(arch) {
			stepError("Can't package arch "+arch+": binary wasn't built. Add linux/"+arch+" to [build]-platforms.", i+1, len(config.Deb.Architectures), 2)
			continue
		}

		err := makeDebPackage(arch)

		if err != nil {
			stepError(err.Error(), i+1, len(config.Deb.Architectures), 2)
		}
	}
}
