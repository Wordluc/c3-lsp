package main

import (
	"flag"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/pherrymason/c3-lsp/internal/c3c"
	"github.com/pherrymason/c3-lsp/internal/lsp/server"
	"github.com/pherrymason/c3-lsp/pkg/option"
)

func cmdLineArguments() (server.ServerOpts, bool, bool, bool) {
	var isTcp = flag.Bool("isTcp", false, "is lsp work on tcp: 127.0.0.0:9696")
	var showHelp = flag.Bool("help", false, "Shows this help")
	var sendCrashReports = flag.Bool("send-crash-reports", false, "Automatically reports crashes to server.")
	var showVersion = flag.Bool("version", false, "Shows server version")

	var logFilePath = flag.String("log-path", "", "Enables logs and sets its filepath")
	var debug = flag.Bool("debug", false, "Enables debug mode")

	// C3 Options
	flag.String("lang-version", "0.6.2", "Specify C3 language version. Deprecated.")
	var c3cPath = flag.String("c3c-path", "", "Path where c3c is located.")
	var stdlibPath = flag.String("stdlib-path", "", "Path to stdlib sources. Allows stdlib inspections.")

	var diagnosticsDelay = flag.Int("diagnostics-delay", 150, "Delay calculation of code diagnostics after modifications in source. In milliseconds, default 2000 ms.")

	flag.Parse()

	c3cPathOpt := option.None[string]()
	if *c3cPath != "" {
		c3cPathOpt = option.Some(*c3cPath)
	}
	stdlibPathOpt := option.None[string]()
	if *stdlibPath != "" {
		stdlibPathOpt = option.Some(*stdlibPath)
	}

	logFilePathOpt := option.None[string]()
	if *logFilePath != "" {
		logFilePathOpt = option.Some(*logFilePath)
	}

	//log.Printf("Version: %s\n", *c3Version)
	//log.Printf("Logpath: %s\n", *logFilePath)
	//log.Printf("Delay: %d\n", *diagnosticsDelay)
	//log.Printf("---------------")

	return server.ServerOpts{
		C3: c3c.C3Opts{
			Version:     option.None[string](),
			Path:        c3cPathOpt,
			StdlibPath:  stdlibPathOpt,
			CompileArgs: []string{},
		},
		Diagnostics: server.DiagnosticsOpts{
			Delay:   time.Duration(*diagnosticsDelay),
			Enabled: true,
		},
		LogFilepath:      logFilePathOpt,
		Debug:            *debug,
		SendCrashReports: *sendCrashReports,
	}, *showHelp, *showVersion, *isTcp
}

func printAppGreet(appName string, version string, commit string) {
	fmt.Printf("%s version %s (%s)\n", appName, version, commit)
}

func printHelp(appName string, version string, commit string) {
	printAppGreet(appName, version, commit)

	fmt.Println("\nOptions")
	flag.PrintDefaults()
}

func buildInfo() string {
	var Commit = func() string {
		if info, ok := debug.ReadBuildInfo(); ok {
			for _, setting := range info.Settings {
				if setting.Key == "vcs.revision" {
					return setting.Value
				}
			}
		}

		return ""
	}()

	return Commit
}
