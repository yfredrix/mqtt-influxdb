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

| Variable               | Description                          | Example Value            |
|------------------------|--------------------------------------|--------------------------|
| `MQTT_SERVER_URL`      | URL of the MQTT broker              | `tcp://localhost:1883`   |
| `MQTT_CLIENT_ID`       | Client ID for the MQTT connection   | `mqtt-client`            |
| `MQTT_TOPIC`           | Topic to subscribe to               | `sensors/temperature`    |
| `MQTT_QOS`             | Quality of Service level            | `1`                      |
| `INFLUXDB_URL`         | URL of the InfluxDB instance        | `http://localhost:8086`  |
| `INFLUXDB_TOKEN`       | Authentication token for InfluxDB   | `your-token`             |
| `INFLUXDB_ORG`         | Organization name in InfluxDB       | `your-org`               |
| `INFLUXDB_BUCKET`      | Bucket name in InfluxDB             | `your-bucket`            |

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