package parser

import (
	"errors"
	"net/url"
	"strings"
)

type VLESSConfig struct {
	ID          string
	Host        string
	Port        string
	Network     string
	Security    string
	SNI         string
	Flow        string
	PublicKey   string
	ShortID     string
	Fingerprint string
	Raw         string
}

func ParseVLESS(vlessURL string) (*VLESSConfig, error) {
	if !strings.HasPrefix(vlessURL, "vless://") {
		return nil, errors.New("невалидный VLESS URL")
	}

	u, err := url.Parse(vlessURL)
	if err != nil {
		return nil, err
	}

	params := u.Query()

	cfg := &VLESSConfig{
		ID:          u.User.Username(),
		Host:        u.Hostname(),
		Port:        u.Port(),
		Network:     params.Get("type"),
		Security:    params.Get("security"),
		SNI:         params.Get("sni"),
		Flow:        params.Get("flow"),
		PublicKey:   params.Get("pbk"),
		ShortID:     params.Get("sid"),
		Fingerprint: params.Get("fp"),
		Raw:         vlessURL,
	}

	if cfg.Port == "" {
		cfg.Port = "443"
	}
	if cfg.Network == "" {
		cfg.Network = "tcp"
	}
	if cfg.Security == "" {
		cfg.Security = "none"
	}

	return cfg, nil
}

func Lines(input string) []string {
	var result []string
	lines := strings.Split(input, "\n")
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l != "" {
			result = append(result, l)
		}
	}
	return result
}
