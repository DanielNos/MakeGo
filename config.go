package main

type ApplicationConfig struct {
	Name        string `toml:"name"`
	Version     string `toml:"version"`
	Description string `toml:"description"`
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
	Apt bool `toml:"apt"`
}

type Config struct {
	Application ApplicationConfig `toml:"application"`
	Build       BuildConfig       `toml:"build"`
	Maintainer  MaintainerConfig  `toml:"maintainer"`
	Package     PackagesConfig    `toml:"package"`
}
