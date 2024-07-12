package main

func goArchToDebArch(architecture string) string {
	switch architecture {
	case "386":
		return "i386"
	default:
		return architecture
	}
}

func goArchToRpmArch(architecture string) string {
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
