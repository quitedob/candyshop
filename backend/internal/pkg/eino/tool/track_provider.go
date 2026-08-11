package tool

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"candypro/api/internal/pkg/shipmenttrack"
)

// seventeenTrackAPI is the 17track v4 batch-tracking endpoint.
const seventeenTrackAPI = "https://api.17track.net/track/v2.2/gettrackinfo"

// SeventeenTrackProvider queries the 17track v4 tracking API. It is env-gated:
// without TRACK_SHIPMENT_17TRACK_API_KEY the tool stays on its deterministic stub.
type SeventeenTrackProvider struct {
	apiKey string
	apiURL string
	http   *http.Client
}

// NewSeventeenTrackProvider creates a provider for the given API token.
func NewSeventeenTrackProvider(apiKey string) *SeventeenTrackProvider {
	return &SeventeenTrackProvider{
		apiKey: apiKey,
		apiURL: seventeenTrackAPI,
		http:   &http.Client{Timeout: 10 * time.Second},
	}
}

// newCarrierTrackerFromEnv returns a 17track provider when the API key is set,
// otherwise nil (the tool falls back to its stub so the agent always answers).
func newCarrierTrackerFromEnv() CarrierTracker {
	key := strings.TrimSpace(os.Getenv("TRACK_SHIPMENT_17TRACK_API_KEY"))
	if key == "" {
		return nil
	}
	return NewSeventeenTrackProvider(key)
}

// Track performs a live 17track lookup. Any provider error is wrapped so the
// caller (NewTrackShipmentTool) can fall back to the deterministic stub.
func (p *SeventeenTrackProvider) Track(ctx context.Context, req *TrackShipmentRequest) (*TrackShipmentResponse, error) {
	payload, err := json.Marshal(map[string]string{
		"tracking_number": strings.TrimSpace(req.BLNumber),
		"carrier_code":    strings.ToUpper(strings.TrimSpace(req.CarrierCode)),
	})
	if err != nil {
		return nil, fmt.Errorf("17track: marshal: %w", err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.apiURL, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("17track: build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("17token", p.apiKey)

	resp, err := p.http.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("17track: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("17track: http %d", resp.StatusCode)
	}

	var parsed struct {
		Code int `json:"code"`
		Data []struct {
			Track struct {
				TrackingNumber     string `json:"tracking_number"`
				LastEvent          string `json:"last_event"`
				DestinationCountry string `json:"destination_country"`
			} `json:"track"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, fmt.Errorf("17track: decode: %w", err)
	}
	if parsed.Code != 0 || len(parsed.Data) == 0 {
		return nil, fmt.Errorf("17track: api code %d", parsed.Code)
	}

	track := parsed.Data[0].Track
	location := strings.TrimSpace(track.DestinationCountry)
	if location == "" {
		location = "UNAVAILABLE"
	}
	lastEvent := strings.TrimSpace(track.LastEvent)
	if lastEvent == "" {
		lastEvent = fmt.Sprintf("No carrier event data available for %s", track.TrackingNumber)
	}
	return &TrackShipmentResponse{
		Status:          shipmenttrack.ResolveTrackingStatus(lastEvent, nil),
		CurrentLocation: location,
		LastEvent:       lastEvent,
		Timestamp:       time.Now().UTC().Format(time.RFC3339),
	}, nil
}
