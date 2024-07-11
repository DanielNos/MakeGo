package main

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
)

func compressSource() error {
	fileName := config.Application.Name + "-" + config.Application.Version
	os.MkdirAll(PKG_TEMP_DIR+"/"+fileName, 0755)

	// Copy source files to temp
	cmd := exec.Command("rsync", "-a",
		".",
		PKG_TEMP_DIR+"/"+fileName,
		"--exclude", "bin", "--exclude", "pkg", "--exclude", ".git", "--exclude", ".vscode", "--exclude", "LICENSE",
	)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return errors.New("Failed to copy source: " + string(output))
	}

	// Compress source files
	cmd = exec.Command("tar",
		"-czvf", "rpmbuild/SOURCES/"+fileName+".tar.gz",
		fileName,
	)
	cmd.Dir = PKG_TEMP_DIR
	output, err = cmd.CombinedOutput()

	if err != nil {
		return errors.New("Failed to compress source: " + string(output))
	}

	return nil
}

func checkRPMRequirments() bool {
	if !isInstalled("rpm") {
		logError(packageIndex-1, "Can't package RPM without rpm installed.")
		return false
	}

	if !isInstalled("tar") {
		logError(packageIndex-1, "Can't package RPM without tar installed.")
		return false
	}

	if !isInstalled("rsync") {
		logError(packageIndex-1, "Can't package RPM without rsync installed.")
		return false
	}

	return true
}

func writeSPECFile(architecture string) {
	file, err := os.Create(PKG_TEMP_DIR + "/rpmbuild/SPECS/" + config.Application.Name + ".spec")
	if err != nil {
		fatal("Failed to create spec file: " + err.Error())
		return
	}
	defer file.Close()

	fileName := config.Application.Name + "-" + config.Application.Version
	writeLine(file, "%global _find_debuginfo_opts %{nil}\n%define debug_package %{nil}\n")

	writeLine(file, "Name: "+config.Application.Name)
	writeLine(file, "Version: "+config.Application.Version)
	writeLine(file, "Release: 1")
	writeLine(file, "Summary: "+config.Application.Description+"\n")

	writeLine(file, "License: "+config.Application.License)
	writeLine(file, "URL: "+config.Application.Url)
	writeLine(file, "Source0: "+fileName+".tar.gz\n")

	writeLine(file, "BuildRequires: golang")
	writeLine(file, "Requires: libc6\n")

	writeLine(file, "%description")
	writeLine(file, config.Application.LongDescription+"\n")

	writeLine(file, "%prep\n%setup\n")

	writeLine(file, "%build")
	writeLine(file, "go get")
	writeLine(file, "go build -o "+fileName+" .\n")

	writeLine(file, "%install")
	writeLine(file, "mkdir -p %{buildroot}/usr/bin/")
	writeLine(file, "install -m 755 "+fileName+" %{buildroot}/usr/bin/"+config.Application.Name+"\n")

	writeLine(file, "%files")
	writeLine(file, "/usr/bin/"+config.Application.Name+"\n")
}

func packageRPM() {
	logStep(packageIndex, packageFormatCount, "Packaging RPM")
	packageIndex++

	// Check requirments
	if !checkRPMRequirments() {
		return
	}

	// Create rpmbuild directories
	rpmbuild := PKG_TEMP_DIR + "/rpmbuild"
	makeDirs([]string{rpmbuild, rpmbuild + "/BUILD", rpmbuild + "/RPMS", rpmbuild + "/SOURCES", rpmbuild + "/SPECS", rpmbuild + "/SRPMS"}, 0755)

	// Compress and prepare source code
	err := compressSource()
	if err != nil {
		logError(packageIndex-1, err.Error())
		return
	}

	// Create SPEC file
	writeSPECFile("linux/amd64")
	absRpmbuild, _ := filepath.Abs("./" + rpmbuild)

	// Package
	cmd := exec.Command("rpmbuild",
		"--without", "debuginfo",
		"--define", "_topdir "+absRpmbuild,
		"-ba", "./rpmbuild/SPECS/"+config.Application.Name+".spec",
	)
	cmd.Dir, _ = filepath.Abs(PKG_TEMP_DIR)

	output, err := cmd.CombinedOutput()
	if err != nil {
		logError(packageIndex-1, "Failed to package: "+string(output))
	}

	// Move packages
	packageName := config.Application.Name + "-" + config.Application.Version + "-1.x86_64.rpm"
	err = os.Rename(PKG_TEMP_DIR+"/rpmbuild/RPMS/x86_64/"+packageName, PKG_DIR+"/"+packageName)
	if err != nil {
		println("ERROR: " + err.Error())
	}

	sourcePackageName := config.Application.Name + "-" + config.Application.Version + "-1.src.rpm"
	err = os.Rename(PKG_TEMP_DIR+"/rpmbuild/SRPMS/"+sourcePackageName, PKG_DIR+"/"+sourcePackageName)
	if err != nil {
		println("ERROR: " + err.Error())
	}

}
