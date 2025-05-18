package main

import (
    "context"
    "encoding/json"
    "log"
    "math"
    "net/http"
    "os"
    "os/signal"
    "sync"
    "syscall"
    "time"
    "github.com/gorilla/websocket"
    "github.com/segmentio/kafka-go"
	"regexp"
    "strings"
    "unicode"
)

// --- Flight Data Types and Cache ---

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

var (
    flightCache = make(map[string]FlightUpdate)
    cacheMutex  = &sync.RWMutex{}
)

// --- WebSocket Server ---

var (
    upgrader  = websocket.Upgrader{
        ReadBufferSize:  1024,
        WriteBufferSize: 1024,
        CheckOrigin:     func(r *http.Request) bool { return true },
    }
    clients   = make(map[*websocket.Conn]bool)
    clientsMu sync.Mutex
)

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Printf("WebSocket upgrade failed: %v", err)
        return
    }
    defer conn.Close()

    clientsMu.Lock()
    clients[conn] = true
    clientsMu.Unlock()
    log.Printf("New client connected: %s", conn.RemoteAddr())

    for {
        if _, _, err := conn.NextReader(); err != nil {
            clientsMu.Lock()
            delete(clients, conn)
            clientsMu.Unlock()
            log.Printf("Client disconnected: %s", conn.RemoteAddr())
            break
        }
    }
}

func sendFlightUpdates() {
    cacheMutex.RLock()
    updates := make([]FlightUpdate, 0, len(flightCache))
    for _, flight := range flightCache {
        updates = append(updates, flight)
    }
    cacheMutex.RUnlock()

    msg, err := json.Marshal(updates)
    if err != nil {
        log.Printf("JSON marshaling error: %v", err)
        return
    }

    clientsMu.Lock()
    for client := range clients {
        if err := client.WriteMessage(websocket.TextMessage, msg); err != nil {
            log.Printf("WebSocket write error: %v", err)
            client.Close()
            delete(clients, client)
        }
    }
    clientsMu.Unlock()
}

func startWebSocketServer() {
    go func() {
        ticker := time.NewTicker(1 * time.Second)
        defer ticker.Stop()
        for range ticker.C {
            sendFlightUpdates()
        }
    }()

    http.HandleFunc("/ws", handleWebSocket)
    log.Println("WebSocket server starting on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}

// --- Kafka Consumer ---

const (
    cacheCleanupInterval = 5 * time.Minute
    flightMaxAge         = 15 * time.Minute
)

func startKafkaConsumer(ctx context.Context, broker, topic string) {
    conn, err := kafka.DialLeader(ctx, "tcp", broker, topic, 0)
    if err != nil {
        log.Fatalf("Failed to connect to Kafka: %v", err)
    }

    partitions, err := conn.ReadPartitions()
    if err != nil {
        log.Fatalf("Failed to get partitions: %v", err)
    }
    log.Printf("Found %d partitions for topic %s", len(partitions), topic)
    conn.Close()

    for _, partition := range partitions {
        log.Printf("Starting reader for partition %d", partition.ID)
        go readPartition(ctx, broker, topic, partition.ID)
    }
}

func readPartition(ctx context.Context, broker, topic string, partition int) {
    r := kafka.NewReader(kafka.ReaderConfig{
        Brokers:     []string{broker},
        Topic:       topic,
        Partition:   partition,
        MinBytes:    1e3,
        MaxBytes:    10e6,
        StartOffset: kafka.LastOffset,
        MaxWait:     1 * time.Second,
    })
    defer r.Close()

    log.Printf("Partition %d reader started", partition)

    for {
        select {
        case <-ctx.Done():
            return
        default:
            m, err := r.ReadMessage(ctx)
            if err != nil {
                if ctx.Err() != nil {
                    return
                }
                log.Printf("Partition %d read error: %v", partition, err)
                time.Sleep(1 * time.Second)
                continue
            }
			raw := cleanJSON(m.Value)

			var update FlightUpdate
			if err := json.Unmarshal(raw, &update); err != nil {
				log.Printf("Unmarshal error after cleaning: %v; raw=%q", err, raw)
				continue
}

            if !isValidCoordinate(update.Latitude, update.Longitude) {
                continue
            }

            cacheMutex.RLock()
            checkCollisions(update)
            cacheMutex.RUnlock()

            cacheMutex.Lock()
            flightCache[update.ICAO24] = update
            cacheMutex.Unlock()
			log.Printf("Cached flight %s at [%f, %f]", update.Callsign, update.Latitude, update.Longitude)

        }
    }
}

func checkCollisions(update FlightUpdate) {
    
}

func isPotentialCollision(a, b FlightUpdate) bool {
    if math.Abs(float64(a.Timestamp-b.Timestamp)) > 60 {
        return false
    }
    horiz := haversine(a.Latitude, a.Longitude, b.Latitude, b.Longitude)
    vert := math.Abs(a.Altitude-b.Altitude) * 3.28084
    return horiz < 5.0 && vert < 1000
}

func haversine(lat1, lon1, lat2, lon2 float64) float64 {
    const R = 6371.0
    dLat := degreesToRadians(lat2 - lat1)
    dLon := degreesToRadians(lon2 - lon1)
    a := math.Sin(dLat/2)*math.Sin(dLat/2) +
        math.Cos(degreesToRadians(lat1))*math.Cos(degreesToRadians(lat2))*
            math.Sin(dLon/2)*math.Sin(dLon/2)
    c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
    return (R * c) * 0.539957
}

func degreesToRadians(d float64) float64 {
    return d * math.Pi / 180
}

func isValidCoordinate(lat, lon float64) bool {
    return lat >= -90 && lat <= 90 && lon >= -180 && lon <= 180
}

func cleanupStaleFlights(ctx context.Context) {
    ticker := time.NewTicker(cacheCleanupInterval)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            now := time.Now().Unix()
            removed := 0
            cacheMutex.Lock()
            for id, flight := range flightCache {
                if now-flight.Timestamp > int64(flightMaxAge.Seconds()) {
                    delete(flightCache, id)
                    removed++
                }
            }
            cacheMutex.Unlock()
            if removed > 0 {
                log.Printf("Cleaned up %d stale flights. Cache size: %d", removed, len(flightCache))
            }
        }
    }
}

func cleanJSON(raw []byte) []byte {
    // 1) Drop any leading BOM or Unicode‑space (e.g. NBSP U+00A0)
    s := string(raw)
    s = strings.TrimLeftFunc(s, func(r rune) bool {
        return unicode.IsSpace(r) || r == '\ufeff' || r == '\u00a0'
    })

    // 2) Normalize the timestamp field if someone accidentally wrapped it in quotes
    //    This regex matches "timestamp": optional spaces, then an optional ', an optional ", digits, optional ", optional '
    //    and rewrites it to "timestamp":<digits>
    tsRe := regexp.MustCompile(`"timestamp"\s*:\s*'?"?(\d+)'?"?`)
    s = tsRe.ReplaceAllString(s, `"timestamp":$1`)

    return []byte(s)
}

// --- Main ---

func main() {
    broker := os.Getenv("KAFKA_BROKER")
    if broker == "" {
        broker = "localhost:9092"
    }
    topic := os.Getenv("KAFKA_TOPIC")
    if topic == "" {
        topic = "flight_updates"
    }

    ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
    defer cancel()

    go cleanupStaleFlights(ctx)
    go startWebSocketServer()
    go startKafkaConsumer(ctx, broker, topic)

    <-ctx.Done()
    log.Println("Shutting down")
}
