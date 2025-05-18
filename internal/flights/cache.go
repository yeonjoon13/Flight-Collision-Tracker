package flights

import (
    "sync"
	"github.com/yeonjoon13/Flight-Collision-Tracker/internal/model"
)

var (
    flightCache = make(map[string]model.FlightUpdate)
    cacheMutex  = &sync.RWMutex{}
)

func GetFlights() []model.FlightUpdate {
    cacheMutex.RLock()
    defer cacheMutex.RUnlock()
    flights := make([]model.FlightUpdate, 0, len(flightCache))
    for _, f := range flightCache {
        flights = append(flights, f)
    }
    return flights
}

func UpdateFlight(f model.FlightUpdate) {
    cacheMutex.Lock()
    flightCache[f.ICAO24] = f
    cacheMutex.Unlock()
}

// ...add other helpers as needed
