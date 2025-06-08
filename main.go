// [hellcat]
package main

import (
	"flag"
	"log"
	"os"

	"hellcat/config"
	"hellcat/parser"
	"hellcat/runner"
	"hellcat/stressor"
)

func main() {
	// Parameters
	vlessURL := flag.String("url", "", "VLESS link")
	listFile := flag.String("list", "", "File containing VLESS links (one per line)")
	threadCount := flag.Int("threads", 50, "Number of stress threads")
	duration := flag.Int("duration", 0, "Duration in seconds (0 = infinite)")
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
		log.Fatal("Please specify --url or --list")
	}

	for _, raw := range urls {
		// Parse VLESS link and generate config
		cfg, err := parser.ParseVLESS(raw)
		if err != nil {
			log.Printf("Parse error: %v", err)
			continue
		}
		confPath := config.Generate(cfg)
		proc := runner.StartXray(confPath)

		// Run stress test: numXray = threadCount
		go stressor.Run("socks5h://127.0.0.1:10808", *threadCount, *duration, *threadCount)

		proc.Wait()
	}
}
