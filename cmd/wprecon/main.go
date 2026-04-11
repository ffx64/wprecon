package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/ffx64/wprecon/internal/app"
)

func main() {
	templatesFlag := flag.String("templates", "", "Templates directory")
	flag.Parse()

	version := "1.0.0"
	fmt.Printf("wprecon v%s - WordPress Reconnaissance Scanner\n", version)

	templatesDir := *templatesFlag
	if templatesDir == "" {
		ex, err := os.Executable()
		if err == nil {
			templatesDir = filepath.Join(filepath.Dir(ex), "templates")
		}

		if _, err := os.Stat(templatesDir); os.IsNotExist(err) {
			templatesDir = "./templates"
		}
	}

	fmt.Printf("Templates directory: %s\n", templatesDir)

	cli, err := app.NewCLIApp(templatesDir)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	if err := cli.Run(os.Args); err != nil {
		log.Fatalf("Error: %v", err)
	}
}
