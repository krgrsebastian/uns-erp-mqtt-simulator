package main

import (
	crand "crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	mochi_mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/hooks/auth"
	"github.com/mochi-mqtt/server/v2/listeners"
)

type ERPData struct {
	OrderID     string `json:"order_id"`
	ProductID   string `json:"product_id"`
	Quantity    int    `json:"quantity"`
	Timestamp   string `json:"timestamp"`
	TimestampMs int64  `json:"timestamp_ms"`
}

type MachineData struct {
	Counter     int     `json:"counter"`
	State       string  `json:"state"`
	Pressure    float64 `json:"pressure"`
	Temperature float64 `json:"temperature"`
	BeltSpeed   float64 `json:"belt_speed"`
	Humidity    float64 `json:"humidity"`
	Timestamp   string  `json:"timestamp"`
	TimestampMs int64   `json:"timestamp_ms"`
}

var (
	currentOrder         ERPData
	currentCounter       int
	currentMachineState  string
	stateDurationCounter int
)

func randomString(n int) string {
	const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, n)
	for i := range result {
		num, err := crand.Int(crand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			panic(err)
		}
		result[i] = letters[num.Int64()]
	}
	return string(result)
}

func generateERPData(baseTime time.Time) ERPData {
	return ERPData{
		OrderID:     fmt.Sprintf("ORD-%s", randomString(4)),
		ProductID:   fmt.Sprintf("PROD-%s", randomString(4)),
		Quantity:    1 + rand.Intn(100),
		Timestamp:   baseTime.Format("2006-01-02 15:04:05"),
		TimestampMs: baseTime.UnixNano() / int64(time.Millisecond),
	}
}

func generateMachineData(baseTime time.Time) MachineData {
	beltSpeed := 0.0
	if currentMachineState == "running" {
		beltSpeed = 1.0 + rand.Float64()*5.0
	}

	return MachineData{
		Counter:     currentCounter,
		State:       currentMachineState,
		Pressure:    10.0 + rand.Float64()*10.0,
		Temperature: 20.0 + rand.Float64()*15.0,
		BeltSpeed:   beltSpeed,
		Humidity:    30.0 + rand.Float64()*20.0,
		Timestamp:   baseTime.Format("2006-01-02 15:04:05"),
		TimestampMs: baseTime.UnixNano() / int64(time.Millisecond),
	}
}

func updateState(changeProbability float64) {
	if stateDurationCounter > 0 && rand.Float64() < changeProbability { // Probability to stay in the current state
		stateDurationCounter--
	} else {
		states := []string{"running", "idle", "unknown", "stop", "maintenance", "cleaning", "inlet jam", "outlet jam"}
		stateWeights := []int{70, 10, 5, 5, 3, 2, 2, 3}
		currentMachineState = states[weightedRandom(stateWeights)]
		stateDurationCounter = rand.Intn(10) // Reset the duration counter
	}
}

func weightedRandom(weights []int) int {
	sum := 0
	for _, weight := range weights {
		sum += weight
	}

	randValue := rand.Intn(sum)
	for i, weight := range weights {
		if randValue < weight {
			return i
		}
		randValue -= weight
	}
	return len(weights) - 1
}

func publish(client mqtt.Client, topic string, payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshaling JSON: %v", err)
		return
	}

	token := client.Publish(topic, 0, false, data)
	token.Wait()
	if token.Error() != nil {
		log.Printf("Error publishing to topic %s: %v", topic, token.Error())
	} else {
		log.Printf("Published to %s: %s", topic, data)
	}
}

func setupClientAndPublish() {
	// Set default values for environment variables
	broker := getEnv("BROKER", "tcp://localhost:1884")
	clientID := getEnv("CLIENT_ID", "mqtt_simulator")
	erpTopic := getEnv("ERP_TOPIC", "umh/v1/umh/cologne/ehrenfeld/office/_historian/erp")
	machineTopicPrefix := getEnv("MACHINE_TOPIC", "umh/v1/umh/cologne/ehrenfeld/office/_historian")
	interval := getEnv("INTERVAL", "5s")

	// Log the environment variables to ensure they are set
	log.Printf("BROKER: %s", broker)
	log.Printf("CLIENT_ID: %s", clientID)
	log.Printf("ERP_TOPIC: %s", erpTopic)
	log.Printf("MACHINE_TOPIC: %s", machineTopicPrefix)
	log.Printf("INTERVAL: %s", interval)

	intervalDuration, err := time.ParseDuration(interval)
	if err != nil {
		log.Fatalf("Error parsing interval duration: %v", err)
	}

	rand.Seed(time.Now().UnixNano()) // Ensure math/rand is seeded

	// MQTT client options
	opts := mqtt.NewClientOptions().AddBroker(broker).SetClientID(clientID)
	client := mqtt.NewClient(opts)

	// Connect to the MQTT broker
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("Error connecting to broker: %v", token.Error())
	}

	log.Printf("Starting MQTT simulator with broker %s, clientID %s, ERP topic %s, Machine topic prefix %s, interval %s",
		broker, clientID, erpTopic, machineTopicPrefix, interval)

	// Initialize the first order and publish it immediately
	startTime := time.Now().Add(-1 * time.Hour) // Change to 1 hour in the past
	currentOrder = generateERPData(startTime)
	currentCounter = 0
	currentMachineState = "running"
	stateDurationCounter = rand.Intn(10)

	erpPayload, err := json.Marshal(currentOrder)
	if err != nil {
		log.Fatalf("Error marshaling initial ERP JSON: %v", err)
	}

	token := client.Publish(erpTopic, 0, false, erpPayload)
	token.Wait()
	log.Printf("Published initial ERP data: %s", erpPayload)

	// Publish historical data
	for t := startTime; t.Before(time.Now()); t = t.Add(intervalDuration) {
		// Update state for historical data period with a lower probability
		updateState(0.95) // 95% chance to stay in the current state

		// Generate Machine data
		machineData := generateMachineData(t)

		// Publish individual data points
		publish(client, fmt.Sprintf("%s/counter", machineTopicPrefix), map[string]interface{}{
			"counter":      machineData.Counter,
			"timestamp_ms": machineData.TimestampMs,
		})
		publish(client, fmt.Sprintf("%s/pressure", machineTopicPrefix), map[string]interface{}{
			"pressure":     machineData.Pressure,
			"timestamp_ms": machineData.TimestampMs,
		})
		publish(client, fmt.Sprintf("%s/temperature", machineTopicPrefix), map[string]interface{}{
			"temperature":  machineData.Temperature,
			"timestamp_ms": machineData.TimestampMs,
		})
		publish(client, fmt.Sprintf("%s/belt_speed", machineTopicPrefix), map[string]interface{}{
			"belt_speed":   machineData.BeltSpeed,
			"timestamp_ms": machineData.TimestampMs,
		})
		publish(client, fmt.Sprintf("%s/humidity", machineTopicPrefix), map[string]interface{}{
			"humidity":     machineData.Humidity,
			"timestamp_ms": machineData.TimestampMs,
		})
		publish(client, fmt.Sprintf("%s/state", machineTopicPrefix), map[string]interface{}{
			"state":        machineData.State,
			"timestamp_ms": machineData.TimestampMs,
		})

		// Increment counter if machine is running
		if machineData.State == "running" {
			currentCounter++
		}

		// Publish new ERP order if counter reaches the target quantity
		if currentCounter >= currentOrder.Quantity {
			currentOrder = generateERPData(t)
			currentCounter = 0

			erpPayload, err := json.Marshal(currentOrder)
			if err != nil {
				log.Fatalf("Error marshaling ERP JSON: %v", err)
			}

			token := client.Publish(erpTopic, 0, false, erpPayload)
			token.Wait()
			log.Printf("Published ERP data: %s", erpPayload)
		}
	}

	// Create tickers
	mainTicker := time.NewTicker(intervalDuration)
	stateTicker := time.NewTicker(20 * time.Second)
	defer mainTicker.Stop()
	defer stateTicker.Stop()

	for {
		select {
		case <-mainTicker.C:
			// Publish ERP data if counter reaches the target quantity
			if currentCounter >= currentOrder.Quantity {
				currentOrder = generateERPData(time.Now())
				currentCounter = 0

				erpPayload, err := json.Marshal(currentOrder)
				if err != nil {
					log.Printf("Error marshaling ERP JSON: %v", err)
					continue
				}

				token := client.Publish(erpTopic, 0, false, erpPayload)
				token.Wait()
				log.Printf("Published ERP data: %s", erpPayload)
			}

			// Generate Machine data
			machineData := generateMachineData(time.Now())

			// Publish individual data points
			publish(client, fmt.Sprintf("%s/counter", machineTopicPrefix), map[string]interface{}{
				"counter":      machineData.Counter,
				"timestamp_ms": machineData.TimestampMs,
			})
			publish(client, fmt.Sprintf("%s/pressure", machineTopicPrefix), map[string]interface{}{
				"pressure":     machineData.Pressure,
				"timestamp_ms": machineData.TimestampMs,
			})
			publish(client, fmt.Sprintf("%s/temperature", machineTopicPrefix), map[string]interface{}{
				"temperature":  machineData.Temperature,
				"timestamp_ms": machineData.TimestampMs,
			})
			publish(client, fmt.Sprintf("%s/belt_speed", machineTopicPrefix), map[string]interface{}{
				"belt_speed":   machineData.BeltSpeed,
				"timestamp_ms": machineData.TimestampMs,
			})
			publish(client, fmt.Sprintf("%s/humidity", machineTopicPrefix), map[string]interface{}{
				"humidity":     machineData.Humidity,
				"timestamp_ms": machineData.TimestampMs,
			})

			// Increment counter if machine is running
			if machineData.State == "running" {
				currentCounter++
			}

		case <-stateTicker.C:
			// Update the state every 20 seconds
			updateState(0.8) // 80% chance to stay in the current state during real-time

			// Generate Machine data with updated state
			machineData := generateMachineData(time.Now())

			// Publish state data
			publish(client, fmt.Sprintf("%s/state", machineTopicPrefix), map[string]interface{}{
				"state":        machineData.State,
				"timestamp_ms": machineData.TimestampMs,
			})

		default:
			time.Sleep(intervalDuration)
		}
	}
}

// getEnv reads an environment variable and returns its value if it exists,
// otherwise it returns the provided default value.
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func main() {
	// Create the new MQTT Server.
	server := mochi_mqtt.New(nil)

	// Allow all connections.
	_ = server.AddHook(new(auth.AllowHook), nil)

	// Create a TCP listener on a standard port.
	tcp := listeners.NewTCP(listeners.Config{ID: "t1", Address: ":1884"})
	err := server.AddListener(tcp)
	if err != nil {
		log.Fatal(err)
	}

	// Serve the MQTT server in a goroutine.
	go func() {
		err := server.Serve()
		if err != nil {
			log.Fatal(err)
		}
	}()

	// Set up signal handling to gracefully shut down the server.
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		done <- true
	}()

	// Call your existing MQTT client setup function here.
	setupClientAndPublish()

	// Run server until interrupted
	<-done

	// Cleanup
	server.Close()
}
