package probe_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/LalatinaHub/common/probe"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProbe_RunConcurrent(t *testing.T) {
	ctx := context.Background()

	// Mock probes
	mockProbe1 := func(ctx context.Context, client *http.Client) probe.Result {
		time.Sleep(10 * time.Millisecond)
		return probe.Result{
			Name:      "Mock-1",
			Passed:    true,
			LatencyMs: 10,
			Region:    "SINGAPORE",
		}
	}

	mockProbe2 := func(ctx context.Context, client *http.Client) probe.Result {
		time.Sleep(15 * time.Millisecond)
		return probe.Result{
			Name:      "Mock-2",
			Passed:    true,
			LatencyMs: 15,
			Country:   "ID",
		}
	}

	mockPanicProbe := func(ctx context.Context, client *http.Client) probe.Result {
		panic("forced panic test")
	}

	results := probe.Run(ctx, http.DefaultClient, mockProbe1, mockProbe2, mockPanicProbe)
	require.Len(t, results, 3)

	assert.Equal(t, "Mock-1", results[0].Name)
	assert.True(t, results[0].Passed)

	assert.Equal(t, "Mock-2", results[1].Name)
	assert.True(t, results[1].Passed)

	assert.False(t, results[2].Passed)
	assert.Contains(t, results[2].Error, "recovered from panic")
}

func TestProbe_YouTubeCDN_MockServer(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("1.2.3.4 => sin01s01-in-f14.1e100.net => sin\n"))
	}))
	defer ts.Close()

	// Custom probe targeting mock server
	mockYouTube := func(ctx context.Context, client *http.Client) probe.Result {
		req, _ := http.NewRequestWithContext(ctx, "GET", ts.URL, nil)
		res, err := client.Do(req)
		if err != nil {
			return probe.Result{Name: "YouTube CDN", Error: err.Error()}
		}
		defer res.Body.Close()

		return probe.Result{
			Name:     "YouTube CDN",
			Passed:   true,
			IATACode: "SIN",
			Region:   "SINGAPORE",
		}
	}

	results := probe.Run(context.Background(), ts.Client(), mockYouTube)
	require.Len(t, results, 1)
	assert.True(t, results[0].Passed)
	assert.Equal(t, "SIN", results[0].IATACode)
	assert.Equal(t, "SINGAPORE", results[0].Region)
}
