# MakeGo

MakeGo is a simple tool for compilation and packaging of Go projects. It uses a single TOML configuration file to automate the build process and package for all major Linux systems.

## How to Build

1. Clone or download the [repository](https://github.com/DanielNos/makego).
2. Build using
    - **MakeGo**: Run `makego`
    - **Go**: Run `go get && go build -ldflags="-w -s" .`

## Installation

MakeGo can be downloaded as a binary, as a package from the [Releases](https://github.com/DanielNos/makego/releases) page or using Go's package manager.

### Installing Using Go

1. Run: `go install github.com/DanielNos/makego@latest`

### Installing Binary

#### Linux

1. Download the correct MakeGo binary for your system. 
2. Move the package to /usr/local/bin: `mv [binary name] /usr/local/bin/makego`

### Windows

1. Download the correct MakeGo binary for your system.
2. Move it to the folder where it should be installed.
3. Open the **Edit the system environment variables** application.
4. Go to the **Advanced** tab and click on the **Environment Variables...** button.
5. Find the **Path** environment variable and click **Edit**.
6. Click **New** and write the path to your installation folder to the new Field.
7. Click **OK**. You may need to restart your terminal or system to apply this change.

### Installing Package

1. Download the correct MakeGo package for your package manager.
2. Install it using the package manager:
    - **APT**: `sudo apt install [package path]`
    - **DNF**: `sudo dnf install [package path]`
    - **PacMan**: `sudo pacman -S [package path]`

## Quick Start

1. Create a new configuration file: `makego new`
2. Edit the make.toml file to suit your needs.
3. Build your project: `makego`

## Commands

### Usage

`makego [action [argument]] [config file]`

### Actions

Default action is `all`.

* `help` - Shows help.
* `new [template]` - Creates a new config from a template. Templates: normal, all, empty
* `clean` or `cln` - Removes all build and package directories.
* `binary` or `bin` - Builds project binaries.
* `package` or `pkg` - Builds project binaries and packages them.
* `all` - Does the same as package.
* `purge` - Removes all build and packaging tools.

### Flags

* `--help` or `-h` - Shows help.
* `--version` or `-v` - Shows version.
* `--time` or `-t` - Prints timestamps for log messages.

## Config File

This is how the default make.toml template looks:

```toml
[application]
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
```

### `application`

|      Field       | Data Type | Description                                                                           |
|------------------|-----------|---------------------------------------------------------------------------------------|
| name             | string    | The name of the application.                                                          |
| version          | string    | The version of the application. Should be either `X.X.X` or `X.X`.                    |
| description      | string    | A short description of your application.                                              |
| long_description | string    | A long description of your application. Used only in RPM packages.                    |
| url              | string    | The url of the main web page for your application.                                    |
| license          | string    | A short name of your project's license (MIT, AGPLv2, ...). Should not contain spaces. |
| gui              | bool      | Whether the application has a GUI or is terminal only.                                |

## `desktop_entry`

|      Field       | Data Type    | Description                                                                                                                                                                                      |
|------------------|--------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| name             | string       | The name of the application used in desktop entries of the application.                                                                                                                          |
| icon             | string       | The icon of the desktop entry.                                                                                                                                                                   |
| categories       | string array | The categories in which your application falls. List of all valid categories can be found at [specifications.freedesktop.org](https://specifications.freedesktop.org/menu-spec/latest/apa.html). |

### `maintainer`

| Field | Data Type | Description                                     |
|-------|-----------|-------------------------------------------------|
| name  | string    | The name and surname of the project maintainer. |
| email | string    | The email address of the project maintainer.    |

### `build`

|   Field   |  Data Type   | Description                                                                                                                                                               |
|-----------|--------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| target    | string       | Build target when running `go build [target]`.                                                                                                                            |
| flags     | string       | Build flags.                                                                                                                                                              |
| platforms | string array | Build platforms in format `[GOOS]/[GOARCH]`. List of all operating systems and architectures can be found on [go.dev/doc](https://go.dev/doc/install/source#environment). |

## `deb`, `pkg`

|     Field     |   Data Type  | Description                                                   |
|---------------|--------------|---------------------------------------------------------------|
| package       | bool         | Should the application be packaged for this packaging system. |
| architectures | string array | Which architectures should be packaged.                       |

## `rpm`

|     Field     |   Data Type  | Description                                                   |
|---------------|--------------|---------------------------------------------------------------|
| package       | bool         | Should the application be packaged for this packaging system. |
| build_src     | bool         | Should the source package be built.                           |
| architectures | string array | Which architectures should be packaged.                       |

## `appimage`

|     Field     |   Data Type  | Description                                                                 |
|---------------|--------------|-----------------------------------------------------------------------------|
| package       | bool         | Should the application be packaged for this packaging system.               |
| architectures | string array | Which architectures should be packaged.                                     |
| custom_apprun | string       | Path to custom AppRun. If left empty, official default AppRun will be used. |

**Supported Architectures:**

| Package Format     | Architectures          |
|--------------------|------------------------|
| deb, RPM, AppImage | amd64, 386, arm, arm64 |
| pkg                | amd64                  |

It's possible to package other architectures that aren't specified here, but they are either unsupported by the packaging system or not tested
