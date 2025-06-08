package runner

import (
	"log"
	"os/exec"
)

func StartXray(configPath string) *exec.Cmd {
	cmd := exec.Command("xray", "-config", configPath)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		buf := make([]byte, 1024)
		for {
			n, _ := stdout.Read(buf)
			if n > 0 {
				log.Printf("[XRAY] %s", buf[:n])
			}
		}
	}()

	go func() {
		buf := make([]byte, 1024)
		for {
			n, _ := stderr.Read(buf)
			if n > 0 {
				log.Printf("[XRAY ERR] %s", buf[:n])
			}
		}
	}()

	if err := cmd.Start(); err != nil {
		log.Fatalf("ERR xray start: %v", err)
	}

	return cmd
}
