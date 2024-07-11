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

type PackagesConfig struct {
	Deb bool `toml:"deb"`
	Rpm bool `toml:"rpm"`
}

type Config struct {
	Application ApplicationConfig `toml:"application"`
	Build       BuildConfig       `toml:"build"`
	Maintainer  MaintainerConfig  `toml:"maintainer"`
	Package     PackagesConfig    `toml:"package"`
}
