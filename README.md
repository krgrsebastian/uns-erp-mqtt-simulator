# MQTT Simulator

This repository contains a simulator for an ERP system with a production machine, using MQTT for communication. The simulator publishes production data and machine status messages to MQTT topics at specified intervals.

## Features

- Generates ERP data with order details and timestamps.
- Simulates machine data including counter, state, pressure, temperature, belt speed, and humidity.
- Publishes machine data at regular intervals and updates the machine state every minute.
- Configurable via environment variables with sensible defaults.

## Getting Started

### Prerequisites

- Go 1.18 or higher
- Docker
- Kubernetes (for deployment)

### Running the Simulator Locally

1. **Clone the Repository:**

    ```bash
    git clone https://github.com/yourusername/mqtt_simulator.git
    cd mqtt_simulator
    ```

2. **Set Environment Variables (Optional):**

    You can set environment variables to configure the simulator. If not set, default values will be used.

    ```bash
    export BROKER="tcp://your-mqtt-broker:1883"
    export CLIENT_ID="mqtt_simulator"
    export ERP_TOPIC="production/erp"
    export MACHINE_TOPIC="production/machine"
    export INTERVAL="5s"
    ```

3. **Run the Simulator:**

    ```bash
    go run main.go
    ```

### Building and Running with Docker

1. **Build the Docker Image:**

    ```bash
    docker build -t yourusername/mqtt_simulator:latest .
    ```

2. **Run the Docker Container:**

    ```bash
    docker run -e BROKER="tcp://your-mqtt-broker:1883" \
               -e CLIENT_ID="mqtt_simulator" \
               -e ERP_TOPIC="production/erp" \
               -e MACHINE_TOPIC="production/machine" \
               -e INTERVAL="5s" \
               yourusername/mqtt_simulator:latest
    ```

### Deploying to Kubernetes

1. **Create the Kubernetes Namespace (if not exists):**

    ```bash
    kubectl create namespace united-manufacturing-hub
    ```

2. **Apply the Deployment:**

    ```bash
    kubectl apply -f deployment.yaml
    ```

3. **Verify the Deployment:**

    ```bash
    kubectl get deployments -n united-manufacturing-hub
    kubectl get pods -n united-manufacturing-hub
    ```

## Configuration

The simulator can be configured using the following environment variables:

- `BROKER`: The MQTT broker URL (default: `tcp://localhost:1883`).
- `CLIENT_ID`: The MQTT client ID (default: `mqtt_simulator`).
- `ERP_TOPIC`: The MQTT topic for ERP data (default: `production/erp`).
- `MACHINE_TOPIC`: The MQTT topic prefix for machine data (default: `production/machine`).
- `INTERVAL`: The interval for publishing machine data (default: `5s`).

## Example Environment Variables

```bash
export BROKER="tcp://your-mqtt-broker:1883"
export CLIENT_ID="mqtt_simulator"
export ERP_TOPIC="production/erp"
export MACHINE_TOPIC="production/machine"
export INTERVAL="5s"
