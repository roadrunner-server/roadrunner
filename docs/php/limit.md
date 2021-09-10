# Embedded Monitoring
RoadRunner is capable of monitoring your application and run soft reset (between requests) if necessary. The previous name - `limit`, current - `supervisor`

## Configuration
Edit your `.rr` file to specify limits for your application:

```yaml
# monitors rr server(s)
http:
  address: "0.0.0.0:8080"
  pool:
    num_workers: 6
    supervisor:
      # watch_tick defines how often to check the state of the workers (seconds)
      watch_tick: 1s
      # ttl defines maximum time worker is allowed to live (seconds)
      ttl: 0
      # idle_ttl defines maximum duration worker can spend in idle mode after first use. Disabled when 0 (seconds)
      idle_ttl: 10s
      # exec_ttl defines maximum lifetime per job (seconds)
      exec_ttl: 10s
      # max_worker_memory limits memory usage per worker (MB)
      max_worker_memory: 100
```