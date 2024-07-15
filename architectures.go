package main

func goArchToPackageArch(architecture string) string {
	switch architecture {
	case "amd64":
		return "x86_64"
	case "386":
		return "i386"
	case "arm":
		return "armhf"
	case "arm64":
		return "aarch64"
	default:
		return architecture
	}
}
