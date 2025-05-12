package model

// FlightUpdate represents a simplified state vector for a single aircraft
type FlightUpdate struct {
    ICAO24    string  `json:"icao24"`
    Callsign  string  `json:"callsign"`
    Origin    string  `json:"origin_country"`
    Latitude  float64 `json:"lat"`
    Longitude float64 `json:"lon"`
    Altitude  float64 `json:"altitude"`
    Velocity  float64 `json:"velocity"`
    Heading   float64 `json:"heading"`
    VertRate  float64 `json:"vertical_rate"`
    Timestamp int64   `json:"timestamp"`
}