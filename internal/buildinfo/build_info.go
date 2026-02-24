package buildinfo

import (
	"fmt"
	"os"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func PrintBuildInfo() {
	version := buildVersion
	if version == "" {
		version = "N/A"
	}
	date := buildDate
	if date == "" {
		date = "N/A"
	}
	commit := buildCommit
	if commit == "" {
		commit = "N/A"
	}

	fmt.Fprintf(os.Stdout, "Build version: %s\n", version)
	fmt.Fprintf(os.Stdout, "Build date: %s\n", date)
	fmt.Fprintf(os.Stdout, "Build commit: %s\n", commit)
}
