package tool

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestSeventeenTrackProvider_MapsResponse verifies a 17track-style response maps
// onto the tool's TrackShipmentResponse.
func TestSeventeenTrackProvider_MapsResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("17token") != "test-token" {
			t.Errorf("missing/incorrect 17token header")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"code": 0,
			"data": [{
				"track": {
					"tracking_number": "MSKU1234567",
					"last_event": "Arrived at destination port",
					"destination_country": "United States"
				}
			}]
		}`))
	}))
	defer srv.Close()

	p := NewSeventeenTrackProvider("test-token")
	p.apiURL = srv.URL

	resp, err := p.Track(context.Background(), &TrackShipmentRequest{BLNumber: "MSKU1234567", CarrierCode: "msku"})
	if err != nil {
		t.Fatalf("Track: %v", err)
	}
	if resp.CurrentLocation != "United States" {
		t.Fatalf("unexpected location %q", resp.CurrentLocation)
	}
	if resp.LastEvent != "Arrived at destination port" {
		t.Fatalf("unexpected last event %q", resp.LastEvent)
	}
	if resp.Status == "" {
		t.Fatal("expected a non-empty status")
	}
	if resp.Timestamp == "" {
		t.Fatal("expected a timestamp")
	}
}

// TestSeventeenTrackProvider_ProviderErrorWrapped verifies a non-200 provider
// response is surfaced as a wrapped error so the tool can fall back to its stub.
func TestSeventeenTrackProvider_ProviderErrorWrapped(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer srv.Close()

	p := NewSeventeenTrackProvider("k")
	p.apiURL = srv.URL

	_, err := p.Track(context.Background(), &TrackShipmentRequest{BLNumber: "MSKU1234567", CarrierCode: "msku"})
	if err == nil {
		t.Fatal("expected error from provider")
	}
	if !strings.Contains(err.Error(), "17track") {
		t.Fatalf("expected wrapped 17track error, got %v", err)
	}
}

// TestNewCarrierTrackerFromEnv_nilWithoutKey verifies the env gate: no API key →
// nil tracker (stub path).
func TestNewCarrierTrackerFromEnv_nilWithoutKey(t *testing.T) {
	t.Setenv("TRACK_SHIPMENT_17TRACK_API_KEY", "")
	if got := newCarrierTrackerFromEnv(); got != nil {
		t.Fatal("expected nil tracker when no API key is set")
	}
}

// TestNewCarrierTrackerFromEnv_providerWithKey verifies the env gate: key set →
// provider.
func TestNewCarrierTrackerFromEnv_providerWithKey(t *testing.T) {
	t.Setenv("TRACK_SHIPMENT_17TRACK_API_KEY", "secret")
	p, ok := newCarrierTrackerFromEnv().(*SeventeenTrackProvider)
	if !ok || p == nil {
		t.Fatal("expected a SeventeenTrackProvider when the API key is set")
	}
}

// TestSeventeenTrackProvider_EmptyData is a sanity check that an empty data array
// yields an error (not a panic).
func TestSeventeenTrackProvider_EmptyData(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = json.Marshal(map[string]any{"code": 0, "data": []any{}})
		_, _ = w.Write([]byte(`{"code": 0, "data": []}`))
	}))
	defer srv.Close()

	p := NewSeventeenTrackProvider("k")
	p.apiURL = srv.URL
	if _, err := p.Track(context.Background(), &TrackShipmentRequest{BLNumber: "MSKU1234567", CarrierCode: "msku"}); err == nil {
		t.Fatal("expected error for empty data")
	}
}
