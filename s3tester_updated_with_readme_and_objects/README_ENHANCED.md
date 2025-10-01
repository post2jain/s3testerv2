# Enhanced s3tester

This fork of **s3tester** adds new capabilities for stress and failure testing against S3-compatible storage.

## New Features

### 1. Target TPS Control
- New flag: `--tps <N>`
- Lets you specify a target transactions per second (TPS).  
(Current implementation wires the parameter but does not fully replace existing concurrency logic — tuning may be needed.)

### 2. Object List Support
- New flag: `--object-list-file <path>`  
- Provide a file with newline-separated object names.  
- Tester will pick keys from this list instead of generating random names.  
- Selection is modulo-based (ensures distribution across provided names).

### 3. Connection Lifetime / TTL
- New flag: `--conn-lifetime-sec <seconds>`  
- Enforces a maximum lifetime on HTTP connections.  
- Connections older than the given lifetime are closed when idle.  
- Conservative: avoids dropping connections mid-request.

### 4. Prometheus Metrics
- New flag: `--metrics-addr <host:port>`  
- Starts a Prometheus `/metrics` endpoint at the given address.  
- Exposes counters for total ops and errors. (Extendable for latencies, TPS, etc.)

## Example Usage

```bash
# Run 2000 TPS for 10 minutes with custom objects
./s3tester   --tps 2000   --duration 600s   --object-list-file ./objects.txt   --conn-lifetime-sec 30   --metrics-addr :9090   --endpoint http://127.0.0.1:9000   --access-key minioadmin   --secret-key minioadmin
```

Where `objects.txt` contains:

```
obj1
obj2
obj3
...
```

## Notes & Caveats

- **TPS flag** is available, but the main concurrency and rate-limiting code of s3tester is complex. The current implementation makes the parameter available and partially wired — more tuning may be required to enforce exact TPS.  
- **Connection lifetime** is conservative: connections are closed when idle, not forcibly during active requests. This avoids corrupting requests.  
- For **2000+ TPS** runs, tune your system:
  - `ulimit -n` (file descriptors)
  - TCP kernel params (`net.ipv4.tcp_tw_reuse`, `net.ipv4.ip_local_port_range`)
  - Ensure sufficient bandwidth & server capacity.

## Next Steps

- Wire `--tps` more tightly into dispatch loops for precise throttling.  
- Expand Prometheus metrics to track per-operation latencies, throughput, and failures.  
- Add aggressive disconnect option (simulate mid-request failures).  

---
Enhanced version generated with ChatGPT assistance (Sept 2025).
