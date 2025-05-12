package collision

import (
    "context"
    "encoding/json"
    "log"

    "github.com/segmentio/kafka-go"
    "github.com/yeonjoon13/Flight-Collision-Tracker/internal/model"
)

// Proximity thresholds
const (
    HorizThreshold = 5.0  // nautical miles
    VertThreshold  = 1000  // feet
)

// CollisionDetector consumes flight updates and detects near-collisions
func RunDetector(ctx context.Context, reader *kafka.Reader) {
    // In-memory store of recent positions keyed by ICAO24
    cache := make(map[string]model.FlightUpdate)

    for {
        m, err := reader.ReadMessage(ctx)
        if err != nil {
            log.Printf("read error: %v", err)
            continue
        }
        var f model.FlightUpdate
        if err := json.Unmarshal(m.Value, &f); err != nil {
            log.Printf("unmarshal error: %v", err)
            continue
        }
        // store new position
        cache[f.ICAO24] = f
        // TODO: compare f against others in cache for collisions
        log.Printf("Processed %s at %.2f,%.2f", f.ICAO24, f.Latitude, f.Longitude)
    }
}