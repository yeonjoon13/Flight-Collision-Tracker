package model

import (
	"encoding/json"
	"time"
)

// FlightUpdate represents a single flight position report
type FlightUpdate struct {
	ICAO24    string  `json:"icao24"`
	Callsign  string  `json:"callsign"`
	Origin    string  `json:"origin_country"`
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"lon"`
	Altitude  float64 `json:"alt_m"`
	Velocity  float64 `json:"velocity"`
	Heading   float64 `json:"heading"`
	VertRate  float64 `json:"vertical_rate"`
	Timestamp int64   `json:"timestamp"`
}

// CollisionAlert represents a potential collision between two aircraft
type CollisionAlert struct {
	Flight1   FlightUpdate `json:"flight1"`
	Flight2   FlightUpdate `json:"flight2"`
	Distance  float64      `json:"distance_nm"`
	AltDiff   float64      `json:"altitude_diff_ft"`
	Timestamp int64        `json:"timestamp"`
}

// UnmarshalFlightUpdate parses a JSON flight update
func UnmarshalFlightUpdate(data []byte, update *FlightUpdate) error {
	if err := json.Unmarshal(data, update); err != nil {
		return err
	}
	
	// If timestamp is missing, set it to current time
	if update.Timestamp == 0 {
		update.Timestamp = time.Now().Unix()
	}
	
	// Clean up callsign
	if update.Callsign != "" {
		// Trim any null bytes or whitespace
		update.Callsign = trimSpace(update.Callsign)
	}
	
	return nil
}

// Helper function to trim whitespace and null bytes
func trimSpace(s string) string {
	// First create a new string without null bytes
	result := ""
	for _, r := range s {
		if r != 0 {
			result += string(r)
		}
	}
	
	// Then try to unquote using JSON unmarshaling
	var unquoted string
	if json.Unmarshal([]byte(`"`+result+`"`), &unquoted) == nil {
		return unquoted
	}
	return result
}