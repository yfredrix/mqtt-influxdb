# MQTT to InfluxDB Bridge

## Overview
This project is a Go-based application that acts as a bridge between an MQTT broker and an InfluxDB instance. It listens to MQTT topics, processes the messages, and stores the data in InfluxDB for further analysis and visualization.

## Features
- Connects to an MQTT broker to subscribe to topics.
- Parses and processes incoming MQTT messages.
- Stores processed data in InfluxDB.
- Configurable via environment variables.
- Includes unit tests for core functionality.

## Prerequisites
- Go 1.18 or later
- Docker (optional, for containerized deployment)
- An MQTT broker (e.g., Mosquitto)
- An InfluxDB instance

## Installation
1. Clone the repository:
   ```bash
   git clone <repository-url>
   cd mqtt-influxdb
   ```
2. Install dependencies:
   ```bash
   go mod tidy
   ```

## Configuration
The application is configured using environment variables. Below are the key variables:

| Variable | Required | Description | Example Value |
|----------|----------|-------------|---------------|
| `MQTTBROKERURL` | Yes | MQTT broker URL | `tcp://localhost:1883` |
| `CLIENTID` | Yes | MQTT client ID | `mqtt-influxdb-bridge` |
| `TOPIC` | Yes | MQTT topic to subscribe to | `sensors/temperature` |
| `QOS` | Yes | MQTT QoS (`0`, `1`, or `2`) | `1` |
| `CAFILE` | Yes | Path to CA certificate file | `/certs/ca.crt` |
| `CERTFILE` | Yes | Path to client certificate file | `/certs/client.crt` |
| `KEYFILE` | Yes | Path to client private key file | `/certs/client.key` |
| `KEEPALIVE` | Yes | MQTT keepalive interval in seconds | `30` |
| `RETRYINTERVAL` | Yes | Reconnect retry interval in milliseconds | `5000` |
| `INFLUXDB_URL` | Yes | InfluxDB server URL | `http://localhost:8086` |
| `INFLUXDB_TOKEN` | Yes | InfluxDB authentication token | `your-token` |
| `INFLUXDB_ORG` | Yes | InfluxDB organization | `your-org` |
| `SESSIONFOLDER` | No | Folder used to persist MQTT session state (empty uses in-memory state) | `/data/session` |
| `DEBUG` | No | Enable Paho/autopaho debug logging (`true`/`false`) | `false` |

### Influx Write Tuning (Optional)

These variables control client-side async batching. If unset, defaults are used.

| Variable                        | Default | Description |
|---------------------------------|---------|-------------|
| `INFLUXDB_WRITE_BATCH_SIZE`     | `5000`  | Maximum points queued before an automatic flush |
| `INFLUXDB_FLUSH_INTERVAL_MS`    | `1000`  | Periodic flush interval in milliseconds |

## Running the Application
1. Set the required environment variables.
2. Run the application:
   ```bash
   go run main.go
   ```

## Testing
Unit tests are included in the `tests/` directory. To run the tests:
```bash
go test ./...
```

## Docker Deployment
A `Dockerfile` is included for containerized deployment. Build and run the Docker image as follows:
```bash
docker build -t mqtt-influxdb .
docker run -d --env-file .env mqtt-influxdb
```

## License
This project is licensed under the MIT License. See the `LICENSE` file for details.