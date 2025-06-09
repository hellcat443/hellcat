// [hellcat]
package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"hellcat/parser"
)

type XrayConfig struct {
	Inbounds  []interface{} `json:"inbounds"`
	Outbounds []interface{} `json:"outbounds"`
}

func Generate(cfg *parser.VLESSConfig) string {
	return GenerateWithPort(cfg, 10808)
}

func GenerateWithPort(cfg *parser.VLESSConfig, port int) string {
	stream := map[string]interface{}{
		"network":  cfg.Network,
		"security": cfg.Security,
	}

	if cfg.Security == "reality" {
		stream["realitySettings"] = map[string]interface{}{
			"serverName":  cfg.SNI,
			"publicKey":   cfg.PublicKey,
			"shortId":     cfg.ShortID,
			"fingerprint": cfg.Fingerprint,
		}
	} else if cfg.Security == "tls" {
		stream["tlsSettings"] = map[string]interface{}{
			"serverName":    cfg.SNI,
			"allowInsecure": true,
		}
	}

	xrayConf := XrayConfig{
		Inbounds: []interface{}{
			map[string]interface{}{
				"port":     port,
				"listen":   "127.0.0.1",
				"protocol": "socks",
				"settings": map[string]interface{}{
					"auth": "noauth",
				},
			},
		},
		Outbounds: []interface{}{
			map[string]interface{}{
				"protocol": "vless",
				"tag":      "vless-out",
				"settings": map[string]interface{}{
					"vnext": []interface{}{
						map[string]interface{}{
							"address": cfg.Host,
							"port":    toInt(cfg.Port),
							"users": []interface{}{
								map[string]interface{}{
									"id":         cfg.ID,
									"encryption": "none",
									"flow":       cfg.Flow,
								},
							},
						},
					},
				},
				"streamSettings": stream,
			},
		},
	}

	fileName := fmt.Sprintf("config_%d_%s.json", port, time.Now().Format("150405"))
	f, err := os.Create(fileName)
	if err != nil {
		log.Fatalf("[hellcat] Error writing config file: %v", err)
	}
	defer f.Close()

	if err := json.NewEncoder(f).Encode(xrayConf); err != nil {
		log.Fatalf("[hellcat] Error encoding config JSON: %v", err)
	}

	return fileName
}

func toInt(s string) int {
	var i int
	fmt.Sscanf(s, "%d", &i)
	return i
}
