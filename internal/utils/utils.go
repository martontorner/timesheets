package utils

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

func GetErrorMessage(err error) string {
	if err == nil {
		return ""
	}

	var messages []string
	for e := err; e != nil; e = errors.Unwrap(e) {
		messages = append(messages, e.Error())
	}

	return strings.Join(messages, ": ")
}

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

func FitArray(a []string, width int) string {
	if width < 2 {
		return "[]"
	}

	remaining := width - 2
	var parts []string

	for _, s := range a {
		partLen := len(s)
		partsLen := len(parts)

		if partsLen > 0 {
			partLen++ // count space if not the first
		}

		if partLen < remaining {
			parts = append(parts, s)
			remaining -= partLen
		} else if partLen > remaining {
			parts = append(parts, "…")
			remaining -= 1
			break
		} else {
			if partsLen < len(a)-1 {
				parts = append(parts, "…")
				remaining -= 1
				break
			} else {
				parts = append(parts, s)
				remaining -= partLen
			}
		}
	}

	return "[" + strings.Join(parts, " ") + "]" + strings.Repeat(" ", remaining)
}

func DefaultString(s string, fallback string) string {
	if s == "" {
		return fallback
	}
	return s
}

func Coalesce[T any](v *T, fallback T) T {
	if v == nil {
		return fallback
	}
	return *v
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
