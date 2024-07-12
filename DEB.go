package main

import (
	"errors"
	"os"
	"os/exec"
)

func checkDEBRequirements() bool {
	if !isInstalled("dpkg-deb") {
		logError(packageIndex-1, "Can't package DEB without dpkg-deb installed.")
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
	writeLine(file, "Architecture: "+goArchToDebArch(arch))
	writeLine(file, "Maintainer: "+config.Maintainer.Name+" <"+config.Maintainer.Email+">")
	writeLine(file, "Description: "+config.Application.Description)
	writeLine(file, "Section: custom")
	writeLine(file, "Priority: optional")
}

func makeDEBPackage(arch string) error {
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

func packageDEB() {
	logStep(packageIndex, packageFormatCount, "Packaging DEB")
	packageIndex++

	// Check requirments
	if !checkDEBRequirements() {
		return
	}

	// Create packagigng directories
	packagingDir := DEB_PKG_DIR + "/" + config.Application.Name + "-" + config.Application.Version
	makeDirs([]string{packagingDir, packagingDir + "/DEBIAN", packagingDir + "/usr/bin"}, 0755)

	// Create packages
	for i, arch := range config.DEB.Architectures {
		logSubStep(i+1, len(config.DEB.Architectures), "Packaging arch "+arch)

		if !isBuildArch(arch) {
			logSubStepError(i+1, len(config.DEB.Architectures), "Can't package arch "+arch+": binary wasn't built. Add linux/"+arch+" to [build]-platforms.")
			continue
		}

		err := makeDEBPackage(arch)

		if err != nil {
			logSubStepError(i+1, len(config.DEB.Architectures), err.Error())
		}
	}
}
