package utils

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

func FitString(s string, width int) string {
	if width < 1 {
		return "…"
	}

	n := len(s)

	switch {
	case n < width:
		return s + strings.Repeat(" ", width-n)
	case n > width:
		return s[:width-1] + "…"
	default:
		return s
	}
}

func DefaultString(s string, fallback string) string {
	if s == "" {
		return fallback
	}
	return s
}

func CreateHttpClient(timeout time.Duration, caPath *string) (*http.Client, error) {
	if caPath == nil {
		return &http.Client{Timeout: timeout}, nil
	}

	caCert, err := os.ReadFile(*caPath)
	if err != nil {
		return nil, err
	}

	caPool := x509.NewCertPool()
	if ok := caPool.AppendCertsFromPEM(caCert); !ok {
		return nil, fmt.Errorf("failed to add CA cert to pool")
	}

	return &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: caPool,
			},
		},
	}, nil
}
