package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// netstatRegex extracts active TCP connections for a specific port
var netstatRegex = regexp.MustCompile(`tcp\s+\d+\s+\d+\s+[\d\.:]+\:(\d+)\s+[\d\.:]+\:\d+\s+(\w+)`)

func main() {
	portFlag := flag.Int("port", 8500, "Port to monitor connections on")
	hostFlag := flag.String("host", "port-opener", "Host running the service")
	flag.Parse()

	fmt.Printf("Monitoring connections on %s:%d\n", *hostFlag, *portFlag)

	count, err := countOpenConnections(*portFlag)
	if err != nil {
		log.Fatalf("Error counting connections: %v\n", err)
	}

	fmt.Printf("# HELP consul_open_http_connections Number of non-TIME_WAIT TCP connections to the specified port\n")
	fmt.Printf("# TYPE consul_open_http_connections gauge\n")
	fmt.Printf("consul_open_http_connections %d\n", count)
}

// countOpenConnections runs netstat and counts TCP connections for the specified port.
func countOpenConnections(port int) (int, error) {
	cmd := exec.Command("netstat", "-tn")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return 0, err
	}

	lines := strings.Split(out.String(), "\n")
	count := 0

	for _, line := range lines {
		matches := netstatRegex.FindStringSubmatch(line)
		if len(matches) < 3 {
			continue
		}

		connPort, _ := strconv.Atoi(matches[1])
		state := matches[2]

		if connPort == port && state != "TIME_WAIT" {
			count++
		}
	}

	return count, nil
}
