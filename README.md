# MQTT Simulator

## Overview

This project is an MQTT simulator that generates and publishes ERP and Machine data to an MQTT broker. It also provides an HTTP API to retrieve the current ERP data. The simulator includes an embedded MQTT broker and client, allowing it to function independently or connect to an external broker.

## Features

- Generates random ERP and Machine data
- Publishes data to MQTT topics
- Provides an HTTP API to retrieve current ERP data
- Configurable through environment variables
- Embedded MQTT broker with authentication hooks

## Requirements

- Go 1.16 or higher
- MQTT broker (embedded or external)
- HTTP server

## Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/your-repo/mqtt-simulator.git
   cd mqtt-simulator
Build the project:

bash
Copy code
go build -o mqtt-simulator
Run the simulator:

bash
Copy code
./mqtt-simulator
Configuration
The simulator can be configured using the following environment variables:

BROKER: MQTT broker URL (default: tcp://localhost:1884)
CLIENT_ID: MQTT client ID (default: mqtt_simulator)
ERP_TOPIC: Topic to publish ERP data (default: umh/v1/umh/cologne/ehrenfeld/office/_historian/erp)
MACHINE_TOPIC: Prefix for machine data topics (default: umh/v1/umh/cologne/ehrenfeld/office/_historian)
INTERVAL: Interval for publishing data (default: 5s)
API_PORT: Port for the HTTP server (default: 8081)
Usage
Running the Simulator
Run the simulator with default settings:

bash
Copy code
./mqtt-simulator
Run the simulator with custom settings:

bash
Copy code
BROKER=tcp://mqtt.example.com:1883 CLIENT_ID=my_simulator ./mqtt-simulator
Accessing the HTTP API
Retrieve the current ERP data:

bash
Copy code
curl http://localhost:8081/erp-data
Code Overview
ERPData Structure
Represents the ERP data with order ID, product ID, quantity, timestamp, and timestamp in milliseconds.

MachineData Structure
Represents the machine data with counter, state, pressure, temperature, belt speed, humidity, timestamp, and timestamp in milliseconds.

AllowHook Structure
An authentication hook that allows connection access for all users and read/write access to all topics.

Main Functions
randomString(n int) string: Generates a random string of specified length.
generateERPData(baseTime time.Time) ERPData: Generates ERP data.
generateMachineData(baseTime time.Time) MachineData: Generates machine data.
updateState(changeProbability float64): Updates the machine state based on probability.
weightedRandom(weights []int) int: Returns an index based on weighted probability.
publish(client mqtt.Client, topic string, payload interface{}): Publishes data to an MQTT topic.
getERPData(w http.ResponseWriter, r *http.Request): HTTP handler to get the current ERP data.
getEnv(key, defaultValue string) string: Reads an environment variable or returns a default value.
Main Routine
Sets default values for environment variables.
Initializes the MQTT broker with an AllowHook.
Connects to the MQTT broker as a client.
Initializes and publishes initial ERP data.
Publishes historical machine data.
Creates tickers for real-time data publishing.
Starts an HTTP server to provide ERP data through an API endpoint.
License
This project is licensed under the MIT License. See the LICENSE file for details.

Kubernetes Deployment
Here is a Kubernetes deployment YAML for the MQTT simulator:

yaml
Copy code
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mqtt-simulator
  labels:
    app: mqtt-simulator
spec:
  replicas: 1
  selector:
    matchLabels:
      app: mqtt-simulator
  template:
    metadata:
      labels:
        app: mqtt-simulator
    spec:
      containers:
      - name: mqtt-simulator
        image: your-docker-image:latest
        ports:
        - containerPort: 8081
        - containerPort: 1884
        env:
        - name: BROKER
          value: "tcp://localhost:1884"
        - name: CLIENT_ID
          value: "mqtt_simulator"
        - name: ERP_TOPIC
          value: "umh/v1/umh/cologne/ehrenfeld/office/_historian/erp"
        - name: MACHINE_TOPIC
          value: "umh/v1/umh/cologne/ehrenfeld/office/_historian"
        - name: INTERVAL
          value: "5s"
        - name: API_PORT
          value: "8081"
---
apiVersion: v1
kind: Service
metadata:
  name: mqtt-simulator
spec:
  selector:
    app: mqtt-simulator
  ports:
  - protocol: TCP
    port: 8081
    targetPort: 8081
  - protocol: TCP
    port: 1884
    targetPort: 1884
  type: NodePort
Replace your-docker-image:latest with the actual Docker image name of your MQTT simulator. This deployment YAML defines a single replica of the MQTT simulator and exposes both the HTTP API (on port 8081) and the MQTT broker (on port 1884) as NodePort services.

To apply this YAML, save it to a file (e.g., mqtt-simulator-deployment.yaml) and run:

bash
Copy code
kubectl apply -f mqtt-simulator-deployment.yaml
This will create the deployment and the service in your Kubernetes cluster.

