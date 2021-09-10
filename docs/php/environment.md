# Environment configuration
All RoadRunner workers will inherit the system configuration available for the parent server process. In addition, you can 
customize the set of env variables to be passed to your workers using part `env` of `.rr` configuration file.

```yaml
server:
  command: "php worker.php"
  env:
     key: value
```

> All keys will be automatically uppercased!

### Default ENV values
RoadRunner provides set of ENV values to help the PHP process to identify how to properly communicate with the server.

Key      | Description
---      | ---
RR_MODE  | Identifies what mode worker should work with ("http", "temporal")
RR_RPC   | Contains RPC connection address when enabled.
RR_RELAY | "pipes" or "tcp://...", depends on server relay configuration.
