package main

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
)

func compressSource() error {
	fileName := config.Application.Name + "-" + config.Application.Version
	os.MkdirAll(RPM_PKG_DIR+"/"+fileName, 0755)

	// Copy source files to temp
	cmd := exec.Command("rsync", "-a",
		".",
		RPM_PKG_DIR+"/"+fileName,
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
	cmd.Dir = RPM_PKG_DIR
	output, err = cmd.CombinedOutput()

	if err != nil {
		return errors.New("Failed to compress source: " + string(output))
	}

	return nil
}

func checkRPMRequirments() bool {
	if !isInstalled("rpm") {
		stepError("Can't package RPM without rpm installed.", packageIndex-1, 2, 1)
		return false
	}

	if !isInstalled("tar") {
		stepError("Can't package RPM without tar installed.", packageIndex-1, 2, 1)
		return false
	}

	if !isInstalled("rsync") {
		stepError("Can't package RPM without rsync installed.", packageIndex-1, 2, 1)
		return false
	}

	return true
}

func writeSPECFile(platform string) {
	goos, goarch := splitPlatArch(platform)

	file, err := os.Create(RPM_PKG_DIR + "/rpmbuild/SPECS/" + config.Application.Name + "-" + goarch + ".spec")
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
	writeLine(file, "Requires: glibc\n")

	writeLine(file, "%description")
	writeLine(file, config.Application.LongDescription+"\n")

	writeLine(file, "%prep\n%setup\n")

	writeLine(file, "%build")
	writeLine(file, "go get")
	writeLine(file, "GOOS="+goos+" GOARCH="+goarch+" go build -o "+fileName+" .\n")

	writeLine(file, "%install")
	writeLine(file, "mkdir -p %{buildroot}/usr/bin/")
	writeLine(file, "install -m 755 "+fileName+" %{buildroot}/usr/bin/"+config.Application.Name+"\n")

	writeLine(file, "%files")
	writeLine(file, "/usr/bin/"+config.Application.Name+"\n")
}

func makeRPMPackage(arch string, buildSource bool) error {
	// Create SPEC file
	writeSPECFile("linux/" + arch)

	// Get absolute rpmbuild path
	absRpmbuild, _ := filepath.Abs("./" + RPM_PKG_DIR + "/rpmbuild")

	// Pick build type flag
	buildFlag := "-bb"
	if buildSource {
		buildFlag = "-bs"
	}

	// Run rpmbuild
	rpmArch := goArchToRpmArch(arch)

	cmd := exec.Command("rpmbuild",
		"--define", "_topdir "+absRpmbuild,
		buildFlag, "./rpmbuild/SPECS/"+config.Application.Name+"-"+arch+".spec",
		"--target", rpmArch,
	)
	cmd.Dir, _ = filepath.Abs(RPM_PKG_DIR)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return errors.New("Failed to package: " + string(output))
	}

	// Move package to package directory
	if buildSource {
		packageName := config.Application.Name + "-" + config.Application.Version + "-1.src.rpm"
		err = os.Rename(RPM_PKG_DIR+"/rpmbuild/SRPMS/"+packageName, PKG_DIR+"/"+packageName)
		if err != nil {
			return errors.New("Failed to move package: " + err.Error())
		}
	} else {
		packageName := config.Application.Name + "-" + config.Application.Version + "-1." + rpmArch + ".rpm"
		err = os.Rename(RPM_PKG_DIR+"/rpmbuild/RPMS/"+rpmArch+"/"+packageName, PKG_DIR+"/"+packageName)
		if err != nil {
			return errors.New("Failed to move package: " + err.Error())
		}
	}

	return nil
}

func packageRPM() {
	step("Packaging RPM", packageIndex, packageFormatCount, 1)
	packageIndex++

	// Check requirments
	if !checkRPMRequirments() {
		return
	}

	// Create rpmbuild directories
	rpmbuild := RPM_PKG_DIR + "/rpmbuild"
	makeDirs([]string{rpmbuild, rpmbuild + "/BUILD", rpmbuild + "/RPMS", rpmbuild + "/SOURCES", rpmbuild + "/SPECS", rpmbuild + "/SRPMS"}, 0755)

	// Compress and prepare source code
	err := compressSource()
	if err != nil {
		stepError(err.Error(), packageIndex-1, packageFormatCount, 2)
		return
	}

	// Create packages
	targetcount := len(config.RPM.Architectures)
	if config.RPM.BuildSource {
		targetcount++
	}

	for i, arch := range config.RPM.Architectures {
		step("Packaging arch "+arch, i+1, targetcount, 2)
		err := makeRPMPackage(arch, false)

		if err != nil {
			stepError(err.Error(), i+1, targetcount, 2)
		}
	}

	// Create source package
	if config.RPM.BuildSource {
		step("Packaging src", targetcount, targetcount, 2)
		err := makeRPMPackage("amd64", true)

		if err != nil {
			stepError(err.Error(), targetcount, targetcount, 2)
		}
	}
}
