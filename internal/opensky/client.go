package opensky

import (
    "context"
    "encoding/json"
    "net/http"
    "time"
    "github.com/yeonjoon13/Flight-Collision-Tracker/internal/model"
)

const defaultURL = "https://opensky-network.org/api/states/all?lamin=24.5&lamax=49.5&lomin=-125&lomax=-66.5"

// OpenSkyResponse matches the API payload
type OpenSkyResponse struct {
    Time   int64             `json:"time"`
    States [][]interface{}   `json:"states"`
}

// FetchStates polls OpenSky and returns a slice of FlightUpdate
func FetchStates(ctx context.Context, url string) ([]model.FlightUpdate, error) {
    if url == "" {
        url = defaultURL
    }
    req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
    client := http.Client{Timeout: 5 * time.Minute}
    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var osResp OpenSkyResponse
    if err := json.NewDecoder(resp.Body).Decode(&osResp); err != nil {
        return nil, err
    }

	getFloat := func(v interface{}) float64 {
        if v == nil {
            return 0
        }
        f, ok := v.(float64)
        if !ok {
            return 0
        }
        return f
    }

    updates := make([]model.FlightUpdate, 0, len(osResp.States))
    for _, s := range osResp.States {
        f := model.FlightUpdate{
            ICAO24:    s[0].(string),
            Callsign:  s[1].(string),
            Origin:    s[2].(string),
            Longitude: getFloat(s[5]),
            Latitude:  getFloat(s[6]),
            Altitude:  getFloat(s[7]),
            Velocity:  getFloat(s[9]),
            Heading:   getFloat(s[10]),
            VertRate:  getFloat(s[11]),
            Timestamp: osResp.Time,
        }
        updates = append(updates, f)
    }
    return updates, nil
}