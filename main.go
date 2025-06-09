// [hellcat]
package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"hellcat/parser"
	"hellcat/stressor"
)

func main() {
	matches, err := filepath.Glob("config*.json")
	if err == nil {
		for _, f := range matches {
			_ = os.Remove(f)
		}
	}

	vlessURL := flag.String("url", "", "VLESS link")
	listFile := flag.String("list", "", "File containing VLESS links")
	threadCount := flag.Int("threads", 50, "Number of stress threads")
	duration := flag.Int("duration", 0, "Duration in seconds (0 = infinite)")
	numXray := flag.Int("instances", 10, "Number of xray-core processes")
	flag.Parse()

	var urls []string

	if *vlessURL != "" {
		urls = append(urls, *vlessURL)
	} else if *listFile != "" {
		data, err := os.ReadFile(*listFile)
		if err != nil {
			log.Fatalf("Failed to read file: %v", err)
		}
		for _, line := range parser.Lines(string(data)) {
			urls = append(urls, line)
		}
	} else {
		log.Fatal("Specify --url or --list")
	}

	for _, raw := range urls {
		cfg, err := parser.ParseVLESS(raw)
		if err != nil {
			log.Printf("Parse error: %v", err)
			continue
		}

		stressor.Run(cfg, *threadCount, *duration, *numXray)
	}
}
