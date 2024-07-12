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

type DEBConfig struct {
	Package       bool     `toml:"package"`
	Architectures []string `toml:"architectures"`
}

type RPMConfig struct {
	Package       bool     `toml:"package"`
	BuildSource   bool     `toml:"build_src"`
	Architectures []string `toml:"architectures"`
}

type Config struct {
	Application ApplicationConfig `toml:"application"`
	Build       BuildConfig       `toml:"build"`
	Maintainer  MaintainerConfig  `toml:"maintainer"`
	DEB         DEBConfig         `toml:"deb"`
	RPM         RPMConfig         `toml:"rpm"`
}
