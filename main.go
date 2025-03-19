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
	defaultConsulPort  = 8500
	defaultUpdateDelay = 1 * time.Second
	metricsFile        = "/workspaces/metrics/tcp_connections.prom" // Prometheus textfile
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

	log.Printf("Starting Prometheus TCP connection exporter:")
	log.Printf(" - Monitoring Port: %d", consulPort)
	log.Printf(" - Update Interval: %s", updateDelay)
}

func main() {
	ticker := time.NewTicker(updateDelay)
	defer ticker.Stop()

	for {
		count := countOpenIPv4Connections(consulPort)
		log.Printf("Open IPv4 connections to port %d: %d\n", consulPort, count)
		writeMetrics(count)
		<-ticker.C
	}
}

// countOpenIPv4Connections runs `netstat -tan` and counts active ESTABLISHED IPv4 connections.
func countOpenIPv4Connections(port int) int {
	cmd := exec.Command("netstat", "-tan")
	out, err := cmd.Output()
	if err != nil {
		log.Printf("Error executing netstat command: %v", err)
		return -1
	}

	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	count := 0

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 6 {
			continue
		}

		state := fields[5]
		remoteAddr := fields[4] // Remote address, should be ":8500"

		// Extract port number from remote address
		extractedPort, err := extractIPv4Port(remoteAddr)
		if err == nil && extractedPort == port && (state != "TIME_WAIT"){
			count++
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error reading netstat output: %v", err)
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

// writeMetrics writes the TCP connection count to the Prometheus metrics file
func writeMetrics(count int) {
	file, err := os.Create(metricsFile)
	if err != nil {
		log.Printf("Error writing to metrics file: %v", err)
		return
	}
	defer file.Close()

	metricContent := fmt.Sprintf("# HELP tcp_connections Number of open TCP connections on port %d\n", consulPort)
	metricContent += "# TYPE tcp_connections gauge\n"
	metricContent += fmt.Sprintf("tcp_connections{port=\"%d\"} %d\n", consulPort, count)

	_, err = file.WriteString(metricContent)
	if err != nil {
		log.Printf("Error writing metric content: %v", err)
	}
}
