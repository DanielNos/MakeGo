package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
)

type ApplicationConfig struct {
	Name            string `toml:"name"`
	Version         string `toml:"version"`
	Url             string `toml:"url"`
	License         string `toml:"license"`
	Description     string `toml:"description"`
	LongDescription string `toml:"long_description"`
	GUI             bool   `toml:"gui"`
}

type DesktopEntryConfig struct {
	Name       string   `toml:"name"`
	IconPath   string   `toml:"icon"`
	Categories []string `toml:"categories"`
}

type BuildConfig struct {
	Target    string   `toml:"target"`
	Flags     string   `toml:"flags"`
	Platforms []string `toml:"platforms"`
}

type MaintainerConfig struct {
	Name  string `toml:"name"`
	Email string `toml:"email"`
}

type SimplePackagingConfig struct {
	Package       bool     `toml:"package"`
	Architectures []string `toml:"architectures"`
}

type PackagingConfig struct {
	Package       bool     `toml:"package"`
	BuildSource   bool     `toml:"build_src"`
	Architectures []string `toml:"architectures"`
}

type AppImagePackagingConfig struct {
	Package       bool     `toml:"package"`
	Architectures []string `toml:"architectures"`
	CustomAppRun  string   `toml:"custom_apprun"`
}

type Config struct {
	Application  ApplicationConfig       `toml:"application"`
	DesktopEntry DesktopEntryConfig      `toml:"desktop_entry"`
	Build        BuildConfig             `toml:"build"`
	Maintainer   MaintainerConfig        `toml:"maintainer"`
	Deb          SimplePackagingConfig   `toml:"deb"`
	RPM          PackagingConfig         `toml:"rpm"`
	Pkg          SimplePackagingConfig   `toml:"pkg"`
	AppImage     AppImagePackagingConfig `toml:"appimage"`
}

func loadConfig() {
	metaData, err := toml.DecodeFile(configFile, &config)

	if err != nil {
		fatal(fmt.Sprintf("Failed to load make config \"%s\": %s", configFile, strings.Split(err.Error(), ":")[1][1:]))
	}

	validateTOML(metaData)
	countPackageFormats()
}

func writeDefaultConfig() {
	// Select config template
	configText := CONFIG_DEFAULT
	if generateTarget == "all" {
		configText = CONFIG_ALL
	} else if generateTarget == "empty" {
		configText = CONFIG_EMPTY
	} else {
		generateTarget = "default"
	}

	// Create file
	info(time.Now(), "Writing config template "+generateTarget+" to "+configFile+".")

	file, err := os.Create(configFile)
	if err != nil {
		fatal("Failed to create config: " + err.Error())
		return
	}
	defer file.Close()

	// Write config
	_, err = file.WriteString(configText)
	if err != nil {
		stepError("Failed to write config: "+err.Error(), 1, packageFormatCount, 0)
		return
	}
}

func validateTOML(metaData toml.MetaData) {
	// Check if all keys are defined
	tomlStructure := [][]string{
		{"application"},
		{"application", "name"},
		{"application", "version"},
		{"application", "url"},
		{"application", "license"},
		{"application", "description"},
		{"application", "long_description"},
		{"application", "gui"},

		{"desktop_entry"},
		{"desktop_entry", "name"},
		{"desktop_entry", "icon"},
		{"desktop_entry", "categories"},

		{"maintainer"},
		{"maintainer", "name"},
		{"maintainer", "email"},

		{"build"},
		{"build", "target"},
		{"build", "flags"},
		{"build", "platforms"},

		{"deb"},
		{"deb", "package"},
		{"deb", "architectures"},

		{"rpm"},
		{"rpm", "package"},
		{"rpm", "build_src"},
		{"rpm", "architectures"},

		{"pkg"},
		{"pkg", "package"},
		{"pkg", "architectures"},

		{"appimage"},
		{"appimage", "package"},
		{"appimage", "architectures"},
		{"appimage", "custom_apprun"},
	}

	for _, key := range tomlStructure {
		if !metaData.IsDefined(key...) {
			errorMessage := "Invalid config \"" + configFile + "\": Missing key " + key[0]

			for i := 1; i < len(key); i++ {
				errorMessage += " - " + key[i]
			}

			fatal(errorMessage)
		}
	}

	// Check if there are additional keys that shouldn't be there
	undecodedKeys := metaData.Undecoded()
	if len(undecodedKeys) > 0 {
		fatal("Invalid config \"" + configFile + "\": Key not found in specification: " + undecodedKeys[0].String())
	}

	// Check if resources exist
	if config.DesktopEntry.IconPath != "" && !fileExists(config.DesktopEntry.IconPath) {
		fatal("Icon file " + config.DesktopEntry.IconPath + " couldn't be found.")
	}
}

const CONFIG_EMPTY = `[application]
name = ""
version = ""
description = ""
long_description = ""
url = ""
license = ""
gui = false

[desktop_entry]
name = ""
icon = ""
categories = [ ]

[maintainer]
name = ""
email = ""

[build]
target = "."
flags = ""
platforms = [ ]

[deb]
package = false
architectures = [ ]

[rpm]
package = false
build_src = false
architectures = [ ]

[pkg]
package = false
architectures = [ ]

[appimage]
package = true
architectures = [ ]
custom_apprun = ""
`

const CONFIG_DEFAULT = `[application]
name = "app"
version = "1.0.0"
description = "My cool application."
long_description = "My cool application."
url = "https://github.com/Username/app"
license = ""
gui = false

[desktop_entry]
name = "App"
icon = "./icon.png"
categories = [ "Utility" ]

[maintainer]
name = "Name Surname"
email = "name.surname@email.com"

[build]
target = "."
flags = "-ldflags=\"-w -s\""
platforms = [ "linux/amd64", "windows/amd64", "darwin/arm64" ]

[deb]
package = false
architectures = [ "amd64" ]

[rpm]
package = false
build_src = true
architectures = [ "amd64" ]

[pkg]
package = true
architectures = [ "amd64" ]

[appimage]
package = true
architectures = [ "amd64" ]
custom_apprun = ""
`

const CONFIG_ALL = `[application]
name = "app"
version = "1.0.0"
description = "My cool application."
long_description = "My cool application."
url = "https://github.com/Username/app"
license = ""
gui = false

[desktop_entry]
name = "App"
icon = "./icon.png"
categories = [ "Utility" ]

[maintainer]
name = "Name Surname"
email = "name.surname@email.com"

[build]
target = "."
flags = "-ldflags=\\"-w -s\\""
platforms = [ "linux/amd64", "linux/386", "linux/arm", "linux/arm64",
"windows/amd64", "windows/386", "windows/arm", "windows/arm64",
"darwin/amd64", "darwin/arm64" ]

[deb]
package = true
architectures = [ "amd64", "386", "arm", "arm64" ]

[rpm]
package = true
build_src = true
architectures = [ "amd64", "386", "arm", "arm64" ]

[pkg]
package = true
architectures = [ "amd64", "386", "arm", "arm64" ]

[appimage]
package = true
architectures = [ "amd64" ]
custom_apprun = ""
`
