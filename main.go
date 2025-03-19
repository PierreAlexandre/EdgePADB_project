package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// Default settings
const (
	defaultConsulPort  = 8500             // Target port as an integer
	defaultUpdateDelay = 1 * time.Second // Update interval
)

var (
	consulPort  int
	updateDelay time.Duration
)

func init() {
	// Use environment variable for port if available
	if portStr, exists := os.LookupEnv("CONSUL_PORT"); exists {
		port, err := strconv.Atoi(portStr)
		if err == nil {
			consulPort = port
		} else {
			log.Printf("Invalid CONSUL_PORT value, using default: %d", defaultConsulPort)
			consulPort = defaultConsulPort
		}
	} else {
		consulPort = defaultConsulPort
	}

	// Use environment variable for update delay if available
	if delayStr, exists := os.LookupEnv("UPDATE_DELAY"); exists {
		delay, err := time.ParseDuration(delayStr)
		if err == nil {
			updateDelay = delay
		} else {
			log.Printf("Invalid UPDATE_DELAY value, using default: %s", defaultUpdateDelay)
			updateDelay = defaultUpdateDelay
		}
	} else {
		updateDelay = defaultUpdateDelay
	}

	log.Printf("Starting connection monitor:")
	log.Printf(" - IPv4-Only Mode")
	log.Printf(" - Monitoring Port: %d", consulPort)
	log.Printf(" - Update Interval: %s", updateDelay)
}

func main() {
	ticker := time.NewTicker(updateDelay)
	defer ticker.Stop()

	for {
		count := countOpenIPv4Connections(consulPort)
		log.Printf("Open IPv4 connections to port %d: %d\n", consulPort, count)
		<-ticker.C
	}
}

// countOpenIPv4Connections runs `ss -tan` and counts active ESTABLISHED IPv4 connections.
func countOpenIPv4Connections(port int) int {
	cmd := exec.Command("ss", "-tan4") // -4 ensures only IPv4 connections
	out, err := cmd.Output()
	if err != nil {
		log.Printf("Error executing ss command: %v", err)
		return -1
	}

	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	count := 0

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 5 { // Ensure we have at least 5 fields
			continue
		}

		state := fields[0]      // e.g., "ESTAB"
		remoteAddr := fields[4] // Remote (peer) address (this is what we need)

		// Extract the port from the remote address (peer)
		extractedPort, err := extractIPv4Port(remoteAddr)
		if err == nil && extractedPort == port {
			if state == "ESTAB" { // Only count established connections
				count++
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error reading ss output: %v", err)
	}

	return count
}


// extractIPv4Port extracts the port number from an IPv4 address string
func extractIPv4Port(address string) (int, error) {
	parts := strings.Split(address, ":")
	if len(parts) < 2 {
		return 0, fmt.Errorf("invalid address format: %s", address)
	}

	portStr := parts[len(parts)-1]
	return strconv.Atoi(portStr)
}
