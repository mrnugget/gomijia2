package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/currantlabs/ble/linux"
)

var (
	configFile = flag.String("config_file", "config.ini", "Config file location")
)

func main() {
	flag.Parse()

	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		log.Printf("[main] Received signal: %s", sig)
		done <- true
	}()

	log.Print("[main] Reading configuration")
	config, err := NewConfig(*configFile)
	if err != nil {
		log.Fatal("Unable to parse configuration")
	}

	log.Print("[main] Starting Linux Device")
	config.Host, err = linux.NewDevice()
	if err != nil {
		log.Fatal(err)
	}

	hostname, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	log.Printf("[main] MQTT broker: %s. Connecting as %s", config.MQTT.Server(), hostname)

	if err := config.MQTT.Connect(hostname); err != nil {
		log.Print("[main] Unable to connect to MQTT broker")
	}

	for _, device := range config.Devices {

		log.Printf("[main:%s] Dialing (%s)", device.Name, device.Addr)
		if err := device.Connect(config.Host); err != nil {
			log.Printf("[main:%s] Failed to connect to device", device.Name)
			continue
		}

		log.Printf("[main:%s] Registering handler", device.Name)
		device.RegisterHandler(config.MQTT)
	}

	// Wait for signal while notification handler respond
	log.Printf("[main] Main goroutine waiting for signal")
	<-done

	for _, device := range config.Devices {
		log.Printf("[main:%s] Disconnecting", device.Name)
		if err := device.Disconnect(); err != nil {
			log.Printf("[main:%s] Failed to disconnect from device", device.Name)
		}
	}

	log.Print("[main] Stopping Linux Device")
	config.Host.Stop()
}
