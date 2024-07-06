package main

func validDebArch(architecture string) bool {
	return architecture == "amd64" || architecture == "arm64" || architecture == "arm" || architecture == "386"
}

func goArchToDebArch(architecture string) string {
	switch architecture {
	case "arm":
		return "armhf"
	case "386":
		return "i386"
	default:
		return architecture
	}
}
