package main

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

func checkPkgRequirements() bool {
	if !isInstalled("pacman") {
		stepError("Can't package pkg without pacman installed.", packageIndex-1, packageFormatCount, 1)
		return false
	}
	return true
}

func makePkgPackage(arch string) error {
	// Write PKGBUILD
	writePKGBUILDFile(arch)

	// Package
	cmd := exec.Command("makepkg")
	cmd.Dir, _ = filepath.Abs(ARCH_PKG_DIR)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return errors.New("Failed to package pkg: " + string(output))
	}

	// Move package
	packageName := config.Application.Name + "-" + config.Application.Version + "-1-" + goArchToPackageArch(arch) + ".pkg.tar.gz"
	err = os.Rename(ARCH_PKG_DIR+"/"+packageName, PKG_DIR+"/"+packageName)

	if err != nil {
		return errors.New("Failed to move package: " + string(output))
	}

	return nil
}

func writePKGBUILDFile(arch string) {
	file, err := os.Create(ARCH_PKG_DIR + "/PKGBUILD")
	if err != nil {
		fatal("Failed to create PKGBUILD file: " + err.Error())
		return
	}
	defer file.Close()

	writeLine(file, "# Maintainer: "+config.Maintainer.Name+" <"+config.Maintainer.Email+">")
	writeLine(file, "pkgname="+config.Application.Name)
	writeLine(file, "pkgver="+config.Application.Version)
	writeLine(file, "pkgrel=1")
	writeLine(file, "pkgdesc=\""+config.Application.Description+"\"")
	writeLine(file, "arch=('"+goArchToPackageArch(arch)+"')")
	writeLine(file, "url=\""+config.Application.Url+"\"")
	writeLine(file, "license=('"+config.Application.License+"')")
	writeLine(file, "source=(\""+config.Application.Name+"-"+config.Application.Version+".tar.gz\")")
	writeLine(file, "sha256sums=('SKIP')\n")

	writeLine(file, "build() {")
	writeLine(file, "   cd \"$srcdir/$pkgname-$pkgver\"")
	writeLine(file, "   GOOS=linux GOARCH="+arch+" go build -o \"$pkgname\" .")
	writeLine(file, "}\n")

	writeLine(file, "package() {")
	writeLine(file, "   cd \"$srcdir/$pkgname-$pkgver\"")
	writeLine(file, "   install -Dm755 \"$pkgname\" \"$pkgdir/usr/bin/$pkgname\"")
	writeLine(file, "}")
}

func packagePkg() {
	step("Packaging pkg", packageIndex, packageFormatCount, 1, false)
	packageIndex++

	// Check requirements
	if !checkPkgRequirements() {
		return
	}

	// Create directories
	makeDirs([]string{ARCH_PKG_DIR + "/SOURCES", ARCH_PKG_DIR + "/SPECS"}, 0755)

	// Copy source
	cmd := exec.Command("cp",
		SRC_PKG_DIR+"/"+config.Application.Name+"-"+config.Application.Version+".tar.gz",
		ARCH_PKG_DIR+"/",
	)
	output, err := cmd.CombinedOutput()

	if err != nil {
		stepError("Failed to copy source: "+string(output), packageIndex-1, packageFormatCount, 1)
		return
	}

	// Create packages
	for i, arch := range config.Pkg.Architectures {
		step("Packaging "+arch, i+1, len(config.Pkg.Architectures), 2, true)

		if runtime.GOARCH != arch {
			stepError("Can't package for architecture "+arch+" on a "+runtime.GOARCH+" system.", i+1, len(config.Deb.Architectures), 2)
			continue
		}

		err := makePkgPackage(arch)

		if err != nil {
			stepError(err.Error(), i+1, len(config.Pkg.Architectures), 2)
		}
	}
}
