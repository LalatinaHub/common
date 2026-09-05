package probe

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/LalatinaHub/common/region"
)

// Result contains the diagnostic result from a network probe.
type Result struct {
	Name      string `json:"name"`
	Passed    bool   `json:"passed"`
	LatencyMs int64  `json:"latency_ms"`
	IATACode  string `json:"iata_code,omitempty"`
	Region    string `json:"region,omitempty"`
	Country   string `json:"country,omitempty"`
	Error     string `json:"error,omitempty"`
}

// ProbeFunc defines the signature for a network diagnostic probe.
type ProbeFunc func(ctx context.Context, client *http.Client) Result

// DefaultProbes returns the standard slice of probes (YouTube CDN & Netflix).
func DefaultProbes() []ProbeFunc {
	return []ProbeFunc{
		YouTubeCDN,
		Netflix,
	}
}

// YouTubeCDN probes Google Video redirector to determine edge server IATA code, city, and latency.
func YouTubeCDN(ctx context.Context, client *http.Client) Result {
	result := Result{
		Name:   "YouTube CDN",
		Passed: false,
	}

	if client == nil {
		client = http.DefaultClient
	}

	req, err := http.NewRequestWithContext(ctx, "GET", "https://redirector.googlevideo.com/report_mapping", nil)
	if err != nil {
		result.Error = err.Error()
		return result
	}

	start := time.Now()
	res, err := client.Do(req)
	result.LatencyMs = time.Since(start).Milliseconds()

	if err != nil {
		result.Error = err.Error()
		return result
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusOK {
		bodyBytes, err := io.ReadAll(res.Body)
		if err != nil {
			result.Error = fmt.Sprintf("read error: %v", err)
			return result
		}

		firstLine := strings.Split(string(bodyBytes), "\n")[0]
		pattern := regexp.MustCompile(`=>\s((\w+-(\w{3}))|(\w{3}))`)
		match := pattern.FindString(firstLine)
		if match != "" {
			parts := strings.Split(match, " ")
			iata := strings.ToUpper(parts[len(parts)-1])
			result.IATACode = iata
			if reg, ok := region.Lookup(iata); ok {
				result.Region = reg
			} else {
				result.Region = iata
			}
			result.Passed = true
			return result
		}

		result.Error = "failed to parse IATA code from response"
		return result
	}

	result.Error = fmt.Sprintf("HTTP status %d", res.StatusCode)
	return result
}

// Netflix probes Netflix service availability and license country unlocking.
func Netflix(ctx context.Context, client *http.Client) Result {
	result := Result{
		Name:   "Netflix",
		Passed: false,
	}

	if client == nil {
		client = http.DefaultClient
	}

	testURLs := []string{
		"https://www.netflix.com/title/81280792",
		"https://www.netflix.com/title/70143836",
	}

	passed := false
	for _, u := range testURLs {
		req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
		if err != nil {
			continue
		}
		req.Header.Set("accept-language", "en-US,en;q=0.9")

		res, err := client.Do(req)
		if err == nil {
			res.Body.Close()
			if res.StatusCode == http.StatusOK {
				passed = true
				break
			}
		}
	}

	if !passed {
		result.Error = "forbidden or streaming restricted"
		return result
	}

	// Geolocation determination via homepage redirect
	start := time.Now()
	homeReq, err := http.NewRequestWithContext(ctx, "GET", "https://www.netflix.com", nil)
	if err != nil {
		result.Error = err.Error()
		return result
	}

	homeRes, err := client.Do(homeReq)
	result.LatencyMs = time.Since(start).Milliseconds()
	if err != nil {
		result.Error = err.Error()
		return result
	}
	defer homeRes.Body.Close()

	if homeRes.StatusCode == http.StatusOK {
		finalURL := homeRes.Request.URL.String()
		pattern := regexp.MustCompile(`com\/(\w{2})`)
		matches := pattern.FindStringSubmatch(finalURL)
		if len(matches) == 2 {
			result.Country = strings.ToUpper(matches[1])
			result.Region = result.Country
		}
		result.Passed = true
		return result
	}

	result.Error = fmt.Sprintf("HTTP status %d", homeRes.StatusCode)
	return result
}

// Run executes a series of probes concurrently using the provided HTTP client.
func Run(ctx context.Context, client *http.Client, probes ...ProbeFunc) []Result {
	if len(probes) == 0 {
		probes = DefaultProbes()
	}

	results := make([]Result, len(probes))
	var wg sync.WaitGroup

	for i, p := range probes {
		wg.Add(1)
		go func(idx int, probeFn ProbeFunc) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					results[idx] = Result{
						Name:   "unknown",
						Passed: false,
						Error:  fmt.Sprintf("recovered from panic: %v", r),
					}
				}
			}()

			results[idx] = probeFn(ctx, client)
		}(i, p)
	}

	wg.Wait()
	return results
}
