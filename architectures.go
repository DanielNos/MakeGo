package main

func goArchToPackageArch(architecture string) string {
	switch architecture {
	case "amd64":
		return "x86_64"
	case "386":
		return "i686"
	case "arm":
		return "armhf"
	case "arm64":
		return "aarch64"
	default:
		return architecture
	}
}

func isStandadtArchitecture(architecture string) bool {
	return architecture == "amd64" || architecture == "386" || architecture == "arm" || architecture == "arm64"
}
