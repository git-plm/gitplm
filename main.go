package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var version = "Development"

func main() {
	initCSV()

	if len(os.Args) < 2 {
		cmdTUI()
		return
	}

	command := os.Args[1]
	args := os.Args[2:]

	switch command {
	case "release":
		cmdRelease(args)
	case "simplify":
		cmdSimplify(args)
	case "combine":
		cmdCombine(args)
	case "http":
		cmdHTTP(args)
	case "update":
		cmdUpdate()
	case "version":
		cmdVersion()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage: %s COMMAND [OPTIONS]\n\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "Commands:\n")
	fmt.Fprintf(os.Stderr, "  (no command)                    Launch interactive TUI\n")
	fmt.Fprintf(os.Stderr, "  release <IPN>                   Process release for IPN\n")
	fmt.Fprintf(os.Stderr, "  simplify <file> -out <file>     Simplify a BOM file\n")
	fmt.Fprintf(os.Stderr, "  combine <file> -out <file>      Combine BOM into output\n")
	fmt.Fprintf(os.Stderr, "  http                            Start KiCad HTTP Library API server\n")
	fmt.Fprintf(os.Stderr, "  update                          Update gitplm to latest version\n")
	fmt.Fprintf(os.Stderr, "  version                         Display version\n")
}

func cmdRelease(args []string) {
	config, err := loadConfig()
	if err != nil {
		log.Printf("Error loading config: %v", err)
		os.Exit(1)
	}

	fs := flag.NewFlagSet("release", flag.ExitOnError)
	flagPMDir := fs.String("pmDir", config.PMDir, "specify location of partmaster CSV files")
	fs.Parse(args)

	if fs.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "Usage: %s release <IPN> [-pmDir <dir>]\n", os.Args[0])
		os.Exit(1)
	}

	releaseIPN := fs.Arg(0)

	updateMsg := CheckForUpdate(version)
	if updateMsg != "" {
		fmt.Println(updateMsg)
	}

	var gLog strings.Builder
	logMsg := func(s string) {
		_, err := gLog.Write([]byte(s))
		if err != nil {
			log.Println("Error writing to gLog: ", err)
		}
		log.Println(s)
	}

	relPath, err := processRelease(releaseIPN, &gLog, *flagPMDir)
	if err != nil {
		logMsg(fmt.Sprintf("release error: %v\n", err))
	} else {
		logMsg(fmt.Sprintf("release %v updated\n", releaseIPN))
	}

	if relPath != "" {
		relIpn := ipn(releaseIPN)
		_, _, _, err := relIpn.parse()
		if err != nil {
			log.Fatal("Error parsing bom IPN: ", err)
		}
		fn := relIpn.base() + ".log"
		logFilePath := filepath.Join(relPath, fn)
		err = os.WriteFile(logFilePath, []byte(gLog.String()), 0644)
		if err != nil {
			log.Println("Error writing log file: ", err)
		}
	}
}

func cmdSimplify(args []string) {
	fs := flag.NewFlagSet("simplify", flag.ExitOnError)
	flagOutput := fs.String("out", "", "output file")
	fs.Parse(args)

	if fs.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "Usage: %s simplify <file> -out <file>\n", os.Args[0])
		os.Exit(1)
	}

	inputFile := fs.Arg(0)

	updateMsg := CheckForUpdate(version)
	if updateMsg != "" {
		fmt.Println(updateMsg)
	}

	in := bom{}
	out := bom{}

	err := loadCSV(inputFile, &in)
	if err != nil {
		log.Printf("Error loading CSV: %v: %v", inputFile, err)
		os.Exit(1)
	}

	for _, l := range in {
		out.addItemMPN(l, true)
	}

	if *flagOutput == "" {
		log.Println("Must specify output file")
		os.Exit(1)
	}

	err = saveCSV(*flagOutput, out)
	if err != nil {
		log.Printf("Error saving CSV: %v: %v", *flagOutput, err)
		os.Exit(1)
	}
}

func cmdCombine(args []string) {
	fs := flag.NewFlagSet("combine", flag.ExitOnError)
	flagOutput := fs.String("out", "", "output file")
	fs.Parse(args)

	if fs.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "Usage: %s combine <file> -out <file>\n", os.Args[0])
		os.Exit(1)
	}

	inputFile := fs.Arg(0)

	updateMsg := CheckForUpdate(version)
	if updateMsg != "" {
		fmt.Println(updateMsg)
	}

	in := bom{}
	out := bom{}

	err := loadCSV(inputFile, &in)
	if err != nil {
		log.Printf("Error loading input CSV: %v: %v", inputFile, err)
		os.Exit(1)
	}

	if fileExists(*flagOutput) {
		err := loadCSV(*flagOutput, &out)
		if err != nil {
			log.Printf("Error loading output CSV: %v: %v", *flagOutput, err)
			os.Exit(1)
		}
	}

	for _, l := range in {
		out.addItemMPN(l, false)
	}

	if *flagOutput == "" {
		log.Println("Must specify output file")
		os.Exit(1)
	}

	err = saveCSV(*flagOutput, out)
	if err != nil {
		log.Printf("Error saving CSV: %v: %v", *flagOutput, err)
		os.Exit(1)
	}
}

func cmdHTTP(args []string) {
	config, err := loadConfig()
	if err != nil {
		log.Printf("Error loading config: %v", err)
		os.Exit(1)
	}

	defaultPort := 7654
	if config.HTTP.Port > 0 {
		defaultPort = config.HTTP.Port
	}
	defaultToken := config.HTTP.Token

	fs := flag.NewFlagSet("http", flag.ExitOnError)
	flagPMDir := fs.String("pmDir", config.PMDir, "specify location of partmaster CSV files")
	flagPort := fs.Int("port", defaultPort, "HTTP server port (default: 7654)")
	flagToken := fs.String("token", defaultToken, "authentication token for HTTP API")
	fs.Parse(args)

	updateMsg := CheckForUpdate(version)
	if updateMsg != "" {
		fmt.Println(updateMsg)
	}

	if *flagPMDir == "" {
		log.Fatal("Error: partmaster directory not specified. Use -pmDir flag or configure gitplm.yml")
	}

	log.Printf("Starting KiCad HTTP Library API server...")
	log.Printf("Partmaster directory: %s", *flagPMDir)
	if *flagToken != "" {
		log.Printf("Authentication enabled with token")
	} else {
		log.Printf("No authentication token specified - server will be open")
	}

	err = StartKiCadServer(*flagPMDir, *flagToken, *flagPort)
	if err != nil {
		log.Fatal("Error starting HTTP server: ", err)
	}
}

func cmdUpdate() {
	updateMsg := CheckForUpdate(version)
	if updateMsg != "" {
		fmt.Println(updateMsg)
	}

	if err := Update(version); err != nil {
		log.Fatalf("Update failed: %v", err)
	}
}

func cmdVersion() {
	if version == "" {
		version = "Development"
	}
	fmt.Printf("%v\n", version)
}

func cmdTUI() {
	config, err := loadConfig()
	if err != nil {
		log.Printf("Error loading config: %v", err)
		os.Exit(1)
	}

	updateMsg := CheckForUpdate(version)

	err = runTUINew(config.PMDir, updateMsg)
	if err != nil {
		log.Fatal("Error running TUI: ", err)
	}
}

func fileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !errors.Is(err, os.ErrNotExist)
}
