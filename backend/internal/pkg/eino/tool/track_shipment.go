package tool

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"candypro/api/internal/pkg/shipmenttrack"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

type TrackShipmentRequest struct {
	BLNumber          string `json:"bl_number" jsonschema_description:"Bill of Lading or AWB tracking number"`
	CarrierCode       string `json:"carrier_code" jsonschema_description:"Shipping line carrier code (e.g., MSKU, COSU)"`
	LastKnownLocation string `json:"last_known_location" jsonschema_description:"Optional latest known location from carrier feed"`
	LastEvent         string `json:"last_event" jsonschema_description:"Optional latest carrier event description"`
	EstimatedArrival  string `json:"estimated_arrival" jsonschema_description:"Optional ETA in RFC3339 or YYYY-MM-DD format"`
}

type TrackShipmentResponse struct {
	Status           string `json:"status"`
	CurrentLocation  string `json:"current_location"`
	EstimatedArrival string `json:"estimated_arrival"`
	LastEvent        string `json:"last_event"`
	Timestamp        string `json:"timestamp"`
}

var trackingNumberPattern = regexp.MustCompile(`^[A-Za-z0-9-]{6,40}$`)

func parseETA(raw string) (time.Time, bool) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return time.Time{}, false
	}

	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return t.UTC(), true
	}
	if t, err := time.Parse("2006-01-02", value); err == nil {
		return t.UTC(), true
	}
	return time.Time{}, false
}

// CarrierTracker performs a live carrier lookup for a B/L or AWB number. A nil
// tracker (or one resolved from env when no API key is set) keeps the tool on the
// deterministic stub so the agent always answers without fabricating data.
type CarrierTracker interface {
	Track(ctx context.Context, req *TrackShipmentRequest) (*TrackShipmentResponse, error)
}

// NewTrackShipmentTool builds the track_shipment tool. When tracker is nil it is
// resolved from the environment (newCarrierTrackerFromEnv); with no API key the
// tool falls back to its deterministic stub behavior.
func NewTrackShipmentTool(ctx context.Context, tracker CarrierTracker) (tool.BaseTool, error) {
	if tracker == nil {
		tracker = newCarrierTrackerFromEnv()
	}
	baseTool, err := utils.InferTool("track_shipment", "Track shipping container or air freight status globally using B/L or AWB.",
		func(ctx context.Context, req *TrackShipmentRequest) (*TrackShipmentResponse, error) {
			if req == nil {
				return nil, fmt.Errorf("request is required")
			}
			blNumber := strings.TrimSpace(req.BLNumber)
			carrierCode := strings.ToUpper(strings.TrimSpace(req.CarrierCode))
			if blNumber == "" || carrierCode == "" {
				return nil, fmt.Errorf("bl_number and carrier_code are required")
			}
			if !trackingNumberPattern.MatchString(blNumber) {
				return nil, fmt.Errorf("invalid bl_number format")
			}
			if len(carrierCode) < 2 || len(carrierCode) > 8 {
				return nil, fmt.Errorf("invalid carrier_code format")
			}

			// Live carrier lookup first; on any provider error fall through to the
			// deterministic stub so the agent still returns a tracking answer.
			if tracker != nil {
				if resp, trackErr := tracker.Track(ctx, req); trackErr == nil && resp != nil {
					return resp, nil
				}
			}

			eta, hasETA := parseETA(req.EstimatedArrival)
			lastEvent := strings.TrimSpace(req.LastEvent)
			if lastEvent == "" {
				lastEvent = fmt.Sprintf("No carrier event data available for %s/%s; provide last_event for deterministic tracking updates.", carrierCode, blNumber)
			}
			location := strings.TrimSpace(req.LastKnownLocation)
			if location == "" {
				location = "UNAVAILABLE"
			}

			var etaPtr *time.Time
			if hasETA {
				etaPtr = &eta
			}

			etaText := ""
			if hasETA {
				etaText = eta.Format("2006-01-02")
			}

			return &TrackShipmentResponse{
				Status:           shipmenttrack.ResolveTrackingStatus(lastEvent, etaPtr),
				CurrentLocation:  location,
				EstimatedArrival: etaText,
				LastEvent:        lastEvent,
				Timestamp:        time.Now().UTC().Format(time.RFC3339),
			}, nil
		})
	if err != nil {
		return nil, err
	}

	return &InvokableReviewEditTool{InvokableTool: baseTool}, nil
}
