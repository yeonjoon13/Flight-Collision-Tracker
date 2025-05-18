import { MapContainer, TileLayer, Marker, Popup, useMap } from 'react-leaflet';
import { useEffect, useState, useRef } from 'react';
import 'leaflet/dist/leaflet.css';
import L from 'leaflet';

// Fix the default marker icon issue in React-Leaflet
delete L.Icon.Default.prototype._getIconUrl;
L.Icon.Default.mergeOptions({
  iconRetinaUrl: 'https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.7.1/images/marker-icon-2x.png',
  iconUrl: 'https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.7.1/images/marker-icon.png',
  shadowUrl: 'https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.7.1/images/marker-shadow.png',
});

function WebSocketManager({ onMessage }) {
  const [status, setStatus] = useState('connecting');
  const wsRef = useRef(null);
  const reconnectTimerRef = useRef(null);

  const connectWebSocket = () => {
    const wsUrl = process.env.REACT_APP_WS_URL || "ws://localhost:8080/ws";

    // Close existing socket if still open
    if (wsRef.current && wsRef.current.readyState !== WebSocket.CLOSED) {
      wsRef.current.close();
    }

    const ws = new WebSocket(wsUrl);
    wsRef.current = ws;

    ws.onopen = () => {
      console.log("✅ WebSocket connected");
      setStatus('connected');
      if (reconnectTimerRef.current) {
        clearTimeout(reconnectTimerRef.current);
        reconnectTimerRef.current = null;
      }
    };

    ws.onmessage = (e) => {
      try {
        const payload = JSON.parse(e.data);

        if (Array.isArray(payload)) {
          payload.forEach(update => {
            if (
              update &&
              typeof update.lat === 'number' &&
              typeof update.lon === 'number'
            ) {
              onMessage(update);
            }
          });
        } else if (
          payload &&
          typeof payload.lat === 'number' &&
          typeof payload.lon === 'number'
        ) {
          // fallback for single‐object messages
          onMessage(payload);
        }
      } catch (err) {
        console.error("Error parsing WebSocket message:", err);
      }
    };

    ws.onerror = (e) => {
      console.error("❌ WebSocket error:", e);
      setStatus('error');
    };

    ws.onclose = () => {
      console.warn("🔌 WebSocket closed");
      setStatus('disconnected');
      // auto‑reconnect in 5s
      reconnectTimerRef.current = setTimeout(() => {
        console.log("Attempting to reconnect...");
        setStatus('connecting');
        connectWebSocket();
      }, 5000);
    };

    // cleanup on unmount
    return () => {
      ws.close();
      if (reconnectTimerRef.current) {
        clearTimeout(reconnectTimerRef.current);
      }
    };
  };

  useEffect(() => {
    const cleanup = connectWebSocket();
    return cleanup;
  }, []);

  return (
    <div className="connection-status">
      Status: {status === 'connected' ? '🟢 Connected' : 
               status === 'connecting' ? '🟡 Connecting...' : 
               status === 'error' ? '🔴 Error' : '🔌 Disconnected'}
    </div>
  );
}

// Auto-center map when flights are visible
function MapRecenter({ flights }) {
  const map = useMap();
  
  useEffect(() => {
    if (flights.length > 0) {
      // Calculate bounds that include all flights
      const bounds = flights.reduce((acc, flight) => {
        return acc.extend([flight.lat, flight.lon]);
      }, L.latLngBounds([]));
      
      // If we have valid bounds, fit the map to them
      if (bounds.isValid()) {
        map.fitBounds(bounds, { padding: [50, 50] });
      }
    }
  }, [flights.length > 0]); // Only recenter when we first get flights
  
  return null;
}

export default function FlightMap() {
  const [flights, setFlights] = useState([]);
  const [lastUpdate, setLastUpdate] = useState(null);

  const handleFlightUpdate = (update) => {
    setFlights(prev => {
      const m = new Map(prev.map(f => [f.icao24, f]));
      m.set(update.icao24, { ...update, lastSeen: Date.now() });
      return Array.from(m.values());
    });
    setLastUpdate(new Date().toLocaleTimeString());
  };
  // Clean up stale flights (not seen in 3 minutes)
  useEffect(() => {
    const STALE_FLIGHT_MS = 3 * 60 * 1000; // 3 minutes
    
    const cleanup = setInterval(() => {
      const now = Date.now();
      setFlights(flights => 
        flights.filter(flight => (now - flight.lastSeen) < STALE_FLIGHT_MS)
      );
    }, 30000); // Run every 30 seconds
    
    return () => clearInterval(cleanup);
  }, []);

  return (
    <div className="flight-tracker">
      <WebSocketManager onMessage={handleFlightUpdate} />
      
      <div className="flight-stats">
        <div>Tracking {flights.length} aircraft</div>
        {lastUpdate && <div>Last update: {lastUpdate}</div>}
      </div>

      

      
      <MapContainer 
        center={[39.5, -98.35]} 
        zoom={4} 
        style={{ height: "calc(100vh - 60px)", width: "100%" }}
      >
        <TileLayer
          attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
          url='https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png'
        />
        
        {flights.map(flight => (
          <Marker 
            key={flight.icao24} 
            position={[flight.lat, flight.lon]}
          >
            <Popup>
              <strong>{flight.callsign || flight.icao24}</strong><br />
              Origin: {flight.origin_country || 'Unknown'}<br />
              Altitude: {Math.round(flight.alt_m)} m ({Math.round(flight.alt_m * 3.28084)} ft)<br />
              {flight.velocity && `Speed: ${Math.round(flight.velocity)} m/s (${Math.round(flight.velocity * 1.94384)} knots)`}<br />
              {flight.heading && `Heading: ${Math.round(flight.heading)}°`}
            </Popup>
          </Marker>
        ))}
        
        <MapRecenter flights={flights} />
      </MapContainer>
    </div>
  );
}