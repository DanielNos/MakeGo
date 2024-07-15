package main

type ApplicationConfig struct {
	Name            string `toml:"name"`
	Version         string `toml:"version"`
	Url             string `toml:"url"`
	License         string `toml:"license"`
	Description     string `toml:"description"`
	LongDescription string `toml:"long_description"`
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

type Config struct {
	Application ApplicationConfig     `toml:"application"`
	Build       BuildConfig           `toml:"build"`
	Maintainer  MaintainerConfig      `toml:"maintainer"`
	Deb         SimplePackagingConfig `toml:"deb"`
	RPM         PackagingConfig       `toml:"rpm"`
	Pkg         SimplePackagingConfig `toml:"pkg"`
}

const CONFIG_EMPTY = "[application]\n" +
	"name = \"\"\n" +
	"version = \"\"\n" +
	"description = \"\"\n" +
	"long_description = \"\"\n" +
	"url = \"\"\n" +
	"license = \"\"\n\n" +

	"[maintainer]\n" +
	"name = \"\"\n" +
	"email = \"\"\n\n" +

	"[build]\n" +
	"target = \".\"\n" +
	"flags = \"\"\n" +
	"platforms = [ ]\n\n" +

	"[deb]\n" +
	"package = false\n" +
	"architectures = [ ]\n\n" +

	"[rpm]\n" +
	"package = false\n" +
	"build_src = false\n" +
	"architectures = [ ]\n\n" +

	"[pkg]\n" +
	"package = false\n" +
	"architectures = [ ]\n"

const CONFIG_DEFAULT = "[application]\n" +
	"name = \"app\"\n" +
	"version = \"1.0.0\"\n" +
	"description = \"My cool application.\"\n" +
	"long_description = \"My cool application.\"\n" +
	"url = \"https://github.com/Username/app\"\n" +
	"license = \"\"\n\n" +

	"[maintainer]\n" +
	"name = \"Name Surname\"\n" +
	"email = \"name.surname@email.com\"\n\n" +

	"[build]\n" +
	"target = \".\"\n" +
	"flags = \"-ldflags=\\\"-w -s\\\"\"\n" +
	"platforms = [ \"linux/amd64\", \"windows/amd64\", \"darwin/arm64\" ]\n\n" +

	"[deb]\n" +
	"package = false\n" +
	"architectures = [ \"amd64\" ]\n\n" +

	"[rpm]\n" +
	"package = false\n" +
	"build_src = true\n" +
	"architectures = [ \"amd64\" ]\n\n" +

	"[pkg]\n" +
	"package = true\n" +
	"architectures = [ \"amd64\" ]\n"

const CONFIG_ALL = "[application]\n" +
	"name = \"app\"\n" +
	"version = \"1.0.0\"\n" +
	"description = \"My cool application.\"\n" +
	"long_description = \"My cool application.\"\n" +
	"url = \"https://github.com/Username/app\"\n" +
	"license = \"\"\n\n" +

	"[maintainer]\n" +
	"name = \"Name Surname\"\n" +
	"email = \"name.surname@email.com\"\n\n" +

	"[build]\n" +
	"target = \".\"\n" +
	"flags = \"-ldflags=\\\"-w -s\\\"\"\n" +
	"platforms = [ \"linux/amd64\", \"linux/386\", \"linux/arm\", \"linux/arm64\",\n" +
	"\"windows/amd64\", \"windows/386\", \"windows/arm\", \"windows/arm64\",\n" +
	"\"darwin/amd64\", \"darwin/arm64\" ]\n\n" +

	"[deb]\n" +
	"package = true\n" +
	"architectures = [ \"amd64\", \"386\", \"arm\", \"arm64\" ]\n\n" +

	"[rpm]\n" +
	"package = true\n" +
	"build_src = true\n" +
	"architectures = [ \"amd64\", \"386\", \"arm\", \"arm64\" ]\n\n" +

	"[pkg]\n" +
	"package = true\n" +
	"architectures = [ \"amd64\", \"386\", \"arm\", \"arm64\" ]\n"
