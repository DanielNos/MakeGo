package main

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

func checkAppImageRequirements() error {
	if !isStandadtArchitecture(runtime.GOARCH) {
		return errors.New("Can't package AppImage on current architecture (" + runtime.GOARCH + "). Supported architectures are: amd64, i686, arm, arm64.")
	}

	if !isInstalled("wget") {
		return errors.New("Can't package AppImage without wget installed.")
	}

	return nil
}

func downloadAppImageTool() error {
	// Check if it's installed
	utilityFileName := "appimagetool-" + goArchToPackageArch(runtime.GOARCH) + ".AppImage"
	if fileExists(UTILITY_DIR + "/" + utilityFileName) {
		return nil
	}

	step("AppImageTool not found, downloading it.", packageIndex-1, packageFormatCount, 1, false)

	// Make directory
	err := os.MkdirAll(UTILITY_DIR, 0755)
	if err != nil {
		return errors.New("Failed to create .makego directory: " + err.Error())
	}

	// Download tool
	cmd := exec.Command("wget",
		"https://github.com/AppImage/AppImageKit/releases/download/continuous/"+utilityFileName,
	)
	cmd.Dir, _ = filepath.Abs(UTILITY_DIR)
	_, err = cmd.CombinedOutput()

	if err != nil {
		return errors.New("Failed to download AppImageTool: " + err.Error())
	}

	// Add execute permission
	err = addXPerm(UTILITY_DIR + "/" + utilityFileName)
	if err != nil {
		return errors.New("Failed to add execute permission to AppImageTool: " + err.Error())
	}

	return nil
}

func writeDesktopEntry(directory string) {
	file, err := os.Create(directory + "/" + config.Application.Name + ".desktop")
	if err != nil {
		stepError("Failed to create desktop file: "+err.Error(), packageIndex-1, packageFormatCount, 1)
	}
	defer file.Close()

	writeLine(file, "[Desktop Entry]")
	writeLine(file, "Name="+config.DesktopEntry.Name)
	writeLine(file, "Exec="+config.Application.Name)
	writeLine(file, "Icon="+config.Application.Name)
	writeLine(file, "Type=Application")

	file.WriteString("Categories=")
	for _, category := range config.DesktopEntry.Categories {
		file.WriteString(category + ";")
	}
	writeLine(file, "")

	if !config.Application.GUI {
		writeLine(file, "Terminal=true")
	}
}

func makeAppImage(stepNumber int, appDir, arch string) error {
	// Check if architecture is supported
	if !isStandadtArchitecture(arch) {
		return errors.New("Can't package AppImage for architecture " + arch + ": unsupported architecture")
	}

	// Download AppRun
	packageArch := goArchToPackageArch(arch)
	if !fileExists(UTILITY_DIR + "/AppRun-" + packageArch) {
		step("AppRun for architecture "+arch+" wasn't found, downloading it.", stepNumber, len(config.AppImage.Architectures), 2, true)

		cmd := exec.Command("wget",
			"https://github.com/AppImage/AppImageKit/releases/download/continuous/AppRun-"+packageArch,
		)
		cmd.Dir, _ = filepath.Abs(UTILITY_DIR)
		_, err := cmd.CombinedOutput()

		if err != nil {
			return errors.New("Failed to download AppRun for " + arch + ": " + err.Error())
		}
	}

	// Copy AppRun
	err := copyFile(UTILITY_DIR+"/AppRun-"+packageArch, appDir+"/AppRun")
	if err != nil {
		return errors.New("Failed to copy AppRun: " + err.Error())
	}

	// Copy binary
	err = copyFile(BIN_DIR+"/"+fileName("linux/amd64"), appDir+"/usr/bin/"+config.Application.Name)
	if err != nil {
		return errors.New("Failed to copy binary: " + err.Error())
	}

	// Package
	cmd := exec.Command(
		"./"+UTILITY_DIR+"/appimagetool-"+goArchToPackageArch(runtime.GOARCH)+".AppImage",
		APPIMAGE_PKG_DIR+"/"+config.Application.Name+".AppDir",
		PKG_DIR+"/"+config.DesktopEntry.Name+"-"+packageArch+".AppImage",
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return errors.New("Failaaed to package binary: " + err.Error() + "\n" + string(output))
	}

	return nil
}

func packageAppImage() {
	step("Packaging AppImage", packageIndex, packageFormatCount, 1, false)
	packageIndex++

	// Check requirements
	err := checkAppImageRequirements()
	if err != nil {
		stepError(err.Error(), packageIndex-1, packageFormatCount, 1)
		return
	}

	// Download AppImage tool
	err = downloadAppImageTool()
	if err != nil {
		stepError(err.Error(), packageIndex-1, packageFormatCount, 1)
	}

	// Create directories
	appDir := APPIMAGE_PKG_DIR + "/" + config.Application.Name + ".AppDir"
	err = os.MkdirAll(appDir+"/usr/bin", 0755)

	if err != nil {
		stepError("Failed to create packaging directories: "+err.Error(), packageIndex-1, packageFormatCount, 1)
		return
	}

	// Create desktop entry
	writeDesktopEntry(appDir)

	// Copy icon
	err = copyFile(config.DesktopEntry.IconPath, appDir+"/"+config.Application.Name+"."+getExtension(config.DesktopEntry.IconPath))
	if err != nil {
		stepError("Failed to copy icon: "+err.Error(), packageIndex-1, packageFormatCount, 1)
		return
	}

	// Package binaries
	for i, arch := range config.AppImage.Architectures {
		step("Packaging "+arch, i+1, len(config.AppImage.Architectures), 2, true)

		err := makeAppImage(i+1, appDir, arch)

		if err != nil {
			stepError(err.Error(), packageIndex-1, packageFormatCount, 2)
		}
	}
}
