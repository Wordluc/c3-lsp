package main

import (
	"fmt"

	"github.com/pherrymason/c3-lsp/internal/lsp/server"
)

const version = "0.4.1"
const prerelease = false
const appName = "C3-LSP"

func main() {
	options, showHelp, showVersion, isTcp := cmdLineArguments()
	commitHash := buildInfo()
	if showHelp {
		printHelp(appName, getLSPVersion(), commitHash)

		return
	}

	if showVersion {
		fmt.Printf("%s\n", version)
		return
	}

	server := server.NewServer(options, appName, version)
	server.Run(isTcp)
}

func getLSPVersion() string {
	if prerelease {
		return version + "-pre"
	}

	return version
}
