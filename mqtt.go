package main

import (
	"fmt"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// MQTT represents an MQTT client configuration
type MQTT struct {
	Host   string
	Port   string
	User   string
	Pass   string
	Client mqtt.Client
	id     string
}

// Server returns a new MQTT connection string
func (m *MQTT) Server() string {
	return fmt.Sprintf("tcp://%s:%s", m.Host, m.Port)
}

// Connect connects to an MQTT broker
func (m *MQTT) Connect(id string) error {
	log.Printf("[MQTT:Connect] Connecting: %s", id)

	m.id = id

	opts := mqtt.NewClientOptions().AddBroker(m.Server())
	opts.SetClientID(id)
	opts.SetKeepAlive(30 * time.Second)
	opts.SetPingTimeout(10 * time.Second)
	if m.User != "" {
		opts.SetUsername(m.User)
	}
	if m.Pass != "" {
		opts.SetPassword(m.Pass)
	}

	log.Print("[MQTT] Creating")
	m.Client = mqtt.NewClient(opts)

	log.Print("[MQTT] Connecting")
	if token := m.Client.Connect(); token.Wait() && token.Error() != nil {
		return (token.Error())
	}
	log.Print("[MQTT] Connected!")

	t := m.Client.Publish("testTopic", 0, false, "test value")
	_ = t.Wait() // Can also use '<-t.Done()' in releases > 1.2.0
	if t.Error() != nil {
		log.Printf("failed to publish test topic: %s", t.Error()) // Use your preferred logging technique (or just fmt.Printf)
	}
	return nil
}

// Disconnect disconnects from an MQTT broker
func (m *MQTT) Disconnect() {
	log.Print("[MQTT] Disconnecting")
	m.Client.Disconnect(250)
}

// Publish publishes a message to an MQTT topic
func (m *MQTT) Publish(name string, r *Reading) {
	log.Printf("[MQTT:Publish] %s (%s)", name, r.String())
	format := "prometheus/job/%s/node/%s/%s"

	// Publish Temperature
	{
		topic := fmt.Sprintf(format, m.id, name, "temperature")
		value := fmt.Sprintf("%04f", r.Temperature)
		m.publish(topic, value)
	}
	// Publish Humidity
	{
		topic := fmt.Sprintf(format, m.id, name, "humidity")
		value := fmt.Sprintf("%04f", r.Humidity)
		m.publish(topic, value)
	}
}
func (m *MQTT) publish(topic string, payload interface{}) {
	log.Printf("[MQTT:publish] %s (%v)", topic, payload)
	m.Client.Publish(topic, 0, false, payload)
}
