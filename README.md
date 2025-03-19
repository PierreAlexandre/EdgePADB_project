# EdgePADB_project

## How to Run

### Prerequisites
- Docker & Docker Compose installed.
- Go installed (if running outside Docker).
- Python 3.7+ (for using `port-opener.py`).

### Running the Service
1. **Using Docker Compose:**
   ```sh
   docker-compose up -d --build
   ```
   This will:
   - Build and start the main go Project.
   - Start the python server with 100 connections on port 8500. (Default)
   - Launch Prometheus for collecting metrics.

2. **Manual Execution:**
   ```sh
   go run .
   ```
   Ensure that the environment variables are set:
   ```sh
   export CONSUL_PORT=8500
   export UPDATE_DELAY=1s
   ```
   The script writes metrics to `/workspaces/metrics/tcp_connections.prom`.

## Package Overview (`main.go`)
The Go program monitors active TCP connections on the port `8500` and exports the count as a Prometheus metric.

## Design Decisions & Trade-offs
- **Data Collection:** Used `netstat -tan` for counting established TCP connections.`netstat` was chosen for simplicity

- **Metrics Export:** Like it was proposed, I used the Prometheus Node Exporter textfile collector method.

- **IPv6 Compatibility:** Currently only IPv4 is supported, but the design allows for easy extension.

- **Handling `TIME_WAIT` connections:** Ignored since Consul does not count them towards the connection limit.

## Test Files & Modifications
- **`port-opener.py`** (provided script): Only the ip address was change.

- **Manual verification:** Metrics can be checked via Prometheus at `http://localhost:9090`.

## Feedback
- **Understanding the Problem:** 
    The problem was indeed well described.
- **Challenges Faced:**
  - Ensuring compatibility with Prometheus Node Exporter.
- **Suggestions for Improvement:**


