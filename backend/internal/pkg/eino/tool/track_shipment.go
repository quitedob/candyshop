package tool

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	einotool "github.com/cloudwego/eino-examples/adk/common/tool"
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

func resolveShipmentStatus(lastEvent string, eta time.Time, hasETA bool) string {
	event := strings.ToLower(strings.TrimSpace(lastEvent))
	if strings.Contains(event, "delivered") || strings.Contains(event, "arrived at destination") {
		return "DELIVERED"
	}
	if strings.Contains(event, "customs hold") || strings.Contains(event, "delay") {
		return "DELAYED"
	}
	if hasETA && time.Now().UTC().After(eta) {
		return "PAST_ETA"
	}
	if event != "" {
		return "IN_TRANSIT"
	}
	return "TRACKING_PENDING"
}

func NewTrackShipmentTool(ctx context.Context) (tool.BaseTool, error) {
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

			eta, hasETA := parseETA(req.EstimatedArrival)
			lastEvent := strings.TrimSpace(req.LastEvent)
			if lastEvent == "" {
				lastEvent = fmt.Sprintf("No carrier event data available for %s/%s; provide last_event for deterministic tracking updates.", carrierCode, blNumber)
			}
			location := strings.TrimSpace(req.LastKnownLocation)
			if location == "" {
				location = "UNAVAILABLE"
			}

			etaText := ""
			if hasETA {
				etaText = eta.Format("2006-01-02")
			}

			return &TrackShipmentResponse{
				Status:           resolveShipmentStatus(lastEvent, eta, hasETA),
				CurrentLocation:  location,
				EstimatedArrival: etaText,
				LastEvent:        lastEvent,
				Timestamp:        time.Now().UTC().Format(time.RFC3339),
			}, nil
		})
	if err != nil {
		return nil, err
	}

	return &einotool.InvokableReviewEditTool{InvokableTool: baseTool}, nil
}
