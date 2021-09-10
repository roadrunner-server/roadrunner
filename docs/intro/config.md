# Configuration

Each of RoadRunner plugins requires proper configuration. By default, such configuration is merged into one file which
must be located in the root of your project. Each service configuration is located under the designated section. The
config file must be named as `.rr.{format}` where the format is `yml`, `json` and others supported by `spf13/viper`.

## Configuration Reference
This is full configuration reference which enables all RoadRunner features.

```yaml
rpc:
  # TCP address:port for listening.
  #
  # Default: "tcp://127.0.0.1:6001"
  listen: tcp://127.0.0.1:6001

# Application server settings (docs: https://roadrunner.dev/docs/php-worker)
server:
  # Worker starting command, with any required arguments.
  #
  # This option is required.
  command: "php psr-worker.php"

  # User name (not UID) for the worker processes. An empty value means to use the RR process user.
  #
  # Default: ""
  user: ""

  # Group name (not GID) for the worker processes. An empty value means to use the RR process user.
  #
  # Default: ""
  group: ""

  # Environment variables for the worker processes.
  #
  # Default: <empty map>
  env:
    - SOME_KEY: "SOME_VALUE"
    - SOME_KEY2: "SOME_VALUE2"

  # Worker relay can be: "pipes", TCP (eg.: tcp://127.0.0.1:6001), or socket (eg.: unix:///var/run/rr.sock).
  #
  # Default: "pipes"
  relay: pipes

  # Timeout for relay connection establishing (only for socket and TCP port relay).
  #
  # Default: 60s
  relay_timeout: 60s

# Logging settings (docs: https://roadrunner.dev/docs/beep-beep-logging)
logs:
  # Logging mode can be "development" or "production". Do not forget to change this value for production environment.
  #
  # Development mode (which makes DPanicLevel logs panic), uses a console encoder, writes to standard error, and
  # disables sampling. Stacktraces are automatically included on logs of WarnLevel and above.
  #
  # Default: "development"
  mode: development

  # Logging level can be "panic", "error", "warn", "info", "debug".
  #
  # Default: "debug"
  level: debug

  # Encoding format can be "console" or "json" (last is preferred for production usage).
  #
  # Default: "console"
  encoding: console

  # Output can be file (eg.: "/var/log/rr_errors.log"), "stderr" or "stdout".
  #
  # Default: "stderr"
  output: stderr

  # Errors only output can be file (eg.: "/var/log/rr_errors.log"), "stderr" or "stdout".
  #
  # Default: "stderr"
  err_output: stderr

  # You can configure each plugin log messages individually (key is plugin name, and value is logging options in same
  # format as above).
  #
  # Default: <empty map>
  channels:
    http:
      mode: development
      level: panic
      encoding: console
      output: stdout
      err_output: stderr
    server:
      mode: production
      level: info
      encoding: json
      output: stdout
      err_output: stdout
    rpc:
      mode: production
      level: debug
      encoding: console
      output: stderr
      err_output: stdout

# Workflow and activity mesh service.
#
# Drop this section for temporal feature disabling.
temporal:
  # Address of temporal server.
  #
  # Default: "127.0.0.1:7233"
  address: 127.0.0.1:7233

  # Activities pool settings.
  activities:
    # How many worker processes will be started. Zero (or nothing) means the number of logical CPUs.
    #
    # Default: 0
    num_workers: 0

    # Maximal count of worker executions. Zero (or nothing) means no limit.
    #
    # Default: 0
    max_jobs: 64

    # Timeout for worker allocation. Zero means no limit.
    #
    # Default: 60s
    allocate_timeout: 60s

    # Timeout for worker destroying before process killing. Zero means no limit.
    #
    # Default: 60s
    destroy_timeout: 60s

    # Supervisor is used to control http workers (previous name was "limit", docs:
    # https://roadrunner.dev/docs/php-limit). "Soft" limits will not interrupt current request processing. "Hard"
    # limit on the contrary - interrupts the execution of the request.
    supervisor:
      # How often to check the state of the workers.
      #
      # Default: 1s
      watch_tick: 1s

      # Maximum time worker is allowed to live (soft limit). Zero means no limit.
      #
      # Default: 0s
      ttl: 0s

      # How long worker can spend in IDLE mode after first using (soft limit). Zero means no limit.
      #
      # Default: 0s
      idle_ttl: 10s

      # Maximal worker memory usage in megabytes (soft limit). Zero means no limit.
      #
      # Default: 0
      max_worker_memory: 128

      # Maximal job lifetime (hard limit). Zero means no limit.
      #
      # Default: 0s
      exec_ttl: 60s

  # Internal temporal communication protocol, can be "proto" or "json".
  #
  # Default: "proto"
  codec: proto

  # Debugging level (only for "json" codec). Set 0 for nothing, 1 for "normal", and 2 for colorized messages.
  #
  # Default: ?
  debug_level: 2

# HTTP plugin settings.
http:
  # Host and port to listen on (eg.: `127.0.0.1:8080`).
  #
  # This option is required.
  address: 127.0.0.1:8080

  # Maximal incoming request size in megabytes. Zero means no limit.
  #
  # Default: 0
  max_request_size: 256

  # Middlewares for the http plugin, order is important. Allowed values is: "headers", "static", "gzip".
  #
  # Default value: []
  middleware: ["headers", "static", "gzip"]

  # Allow incoming requests only from the following subnets (https://en.wikipedia.org/wiki/Reserved_IP_addresses).
  #
  # Default: ["10.0.0.0/8", "127.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16",  "::1/128", "fc00::/7", "fe80::/10"]
  trusted_subnets: [
      "10.0.0.0/8",
      "127.0.0.0/8",
      "172.16.0.0/12",
      "192.168.0.0/16",
      "::1/128",
      "fc00::/7",
      "fe80::/10",
  ]

  # File uploading settings.
  uploads:
    # Directory for file uploads. Empty value means to use $TEMP based on your OS.
    #
    # Default: ""
    dir: "/tmp"

    # Deny files with the following extensions to upload.
    #
    # Default: [".php", ".exe", ".bat"]
    forbid: [".php", ".exe", ".bat", ".sh"]

  # Settings for "headers" middleware (docs: https://roadrunner.dev/docs/http-headers).
  headers:
    # Allows to control CORS headers. Additional headers "Vary: Origin", "Vary: Access-Control-Request-Method",
    # "Vary: Access-Control-Request-Headers" will be added to the server responses. Drop this section for this
    # feature disabling.
    cors:
      # Controls "Access-Control-Allow-Origin" header value (docs: https://mzl.la/2OgD4Qf).
      #
      # Default: ""
      allowed_origin: "*"

      # Controls "Access-Control-Allow-Headers" header value (docs: https://mzl.la/2OzDVvk).
      #
      # Default: ""
      allowed_headers: "*"

      # Controls "Access-Control-Allow-Methods" header value (docs: https://mzl.la/3lbwyXf).
      #
      # Default: ""
      allowed_methods: "GET,POST,PUT,DELETE"

      # Controls "Access-Control-Allow-Credentials" header value (docs: https://mzl.la/3ekJGaY).
      #
      # Default: false
      allow_credentials: true

      # Controls "Access-Control-Expose-Headers" header value (docs: https://mzl.la/3qAqgkF).
      #
      # Default: ""
      exposed_headers: "Cache-Control,Content-Language,Content-Type,Expires,Last-Modified,Pragma"

      # Controls "Access-Control-Max-Age" header value in seconds (docs: https://mzl.la/2PCSdvt).
      #
      # Default: 0
      max_age: 600

    # Automatically add headers to every request passed to PHP.
    #
    # Default: <empty map>
    request:
      input: "custom-header"

    # Automatically add headers to every response.
    #
    # Default: <empty map>
    response:
      X-Powered-By: "RoadRunner"

  # Settings for "static" middleware (docs: https://roadrunner.dev/docs/http-static).
  static:
    # Path to the directory with static assets.
    #
    # This option is required.
    dir: "/path/to/directory"

    # File extensions to forbid.
    #
    # Default: []
    forbid: [".htaccess"]

    # Automatically add headers to every request.
    #
    # Default: <empty map>
    request:
      input: "custom-header"

    # Automatically add headers to every response.
    #
    # Default: <empty map>
    response:
      X-Powered-By: "RoadRunner"

  # Workers pool settings.
  pool:
    # How many worker processes will be started. Zero (or nothing) means the number of logical CPUs.
    #
    # Default: 0
    num_workers: 0

    # Maximal count of worker executions. Zero (or nothing) means no limit.
    #
    # Default: 0
    max_jobs: 64

    # Timeout for worker allocation. Zero means no limit.
    #
    # Default: 60s
    allocate_timeout: 60s

    # Timeout for worker destroying before process killing. Zero means no limit.
    #
    # Default: 60s
    destroy_timeout: 60s

    # Supervisor is used to control http workers (previous name was "limit", docs:
    # https://roadrunner.dev/docs/php-limit). "Soft" limits will not interrupt current request processing. "Hard"
    # limit on the contrary - interrupts the execution of the request.
    supervisor:
      # How often to check the state of the workers.
      #
      # Default: 1s
      watch_tick: 1s

      # Maximum time worker is allowed to live (soft limit). Zero means no limit.
      #
      # Default: 0s
      ttl: 0s

      # How long worker can spend in IDLE mode after first using (soft limit). Zero means no limit.
      #
      # Default: 0s
      idle_ttl: 10s

      # Maximal worker memory usage in megabytes (soft limit). Zero means no limit.
      #
      # Default: 0
      max_worker_memory: 128

      # Maximal job lifetime (hard limit). Zero means no limit.
      #
      # Default: 0s
      exec_ttl: 60s

  # SSL (Secure Sockets Layer) settings (docs: https://roadrunner.dev/docs/http-https).
  ssl:
    # Host and port to listen on (eg.: `127.0.0.1:443`).
    #
    # Default: ":443"
    address: "127.0.0.1:443"

    # Automatic redirect from http:// to https:// schema.
    #
    # Default: false
    redirect: true

    # Path to the cert file. This option is required for SSL working.
    #
    # This option is required.
    cert: /ssl/server.crt

    # Path to the cert key file.
    #
    # This option is required.
    key: /ssl/server.key

    # Path to the root certificate authority file.
    #
    # This option is optional.
    root_ca: /ssl/root.crt

  # FastCGI frontend support.
  fcgi:
    # FastCGI connection DSN. Supported TCP and Unix sockets. An empty value disables this.
    #
    # Default: ""
    address: tcp://0.0.0.0:7921

  # HTTP/2 settings.
  http2:
    # HTTP/2 over non-encrypted TCP connection using H2C.
    #
    # Default: false
    h2c: false

    # Maximal concurrent streams count.
    #
    # Default: 128
    max_concurrent_streams: 128

# Application metrics in Prometheus format (docs: https://roadrunner.dev/docs/beep-beep-metrics). Drop this section
# for this feature disabling.
metrics:
  # Prometheus client address (path /metrics added automatically).
  #
  # Default: "127.0.0.1:2112"
  address: "127.0.0.1:2112"

  # Application-specific metrics (published using an RPC connection to the server).
  collect:
    app_metric:
      type: histogram
      help: "Custom application metric"
      labels: ["type"]
      buckets: [0.1, 0.2, 0.3, 1.0]
      # Objectives defines the quantile rank estimates with their respective absolute error (for summary only).
      objectives:
        - 1.4: 2.3
        - 2.0: 1.4

# Health check endpoint (docs: https://roadrunner.dev/docs/beep-beep-health). If response code is 200 - it means at
# least one worker ready to serve requests. 500 - there are no workers ready to service requests.
# Drop this section for this feature disabling.
status:
  # Host and port to listen on (eg.: `127.0.0.1:2114`). Use the following URL: http://127.0.0.1:2114/health?plugin=http
  # Multiple plugins must be separated using "&" - http://127.0.0.1:2114/health?plugin=http&plugin=rpc where "http" and
  # "rpc" are active (connected) plugins.
  #
  # This option is required.
  address: 127.0.0.1:2114
  
  # Response status code if a requested plugin not ready to handle requests
  # Valid for both /health and /ready endpoints
  #
  # Default: 503
  unavailable_status_code: 503

# Automatically detect PHP file changes and reload connected services (docs:
# https://roadrunner.dev/docs/beep-beep-reload). Drop this section for this feature disabling.
reload:
  # Sync interval.
  #
  # Default: "1s"
  interval: 1s

  # Global patterns to sync.
  #
  # Default: [".php"]
  patterns: [".php"]

  # List of included for sync services (this is a map, where key name is a plugin name).
  #
  # Default: <empty map>
  services:
    http:
      # Directories to sync. If recursive is set to true, recursive sync will be applied only to the directories in
      # "dirs" section. Dot (.) means "current working directory".
      #
      # Default: []
      dirs: ["."]

      # Recursive search for file patterns to add.
      #
      # Default: false
      recursive: true

      # Ignored folders.
      #
      # Default: []
      ignore: ["vendor"]

      # Service specific file pattens to sync.
      #
      # Default: []
      patterns: [".php", ".go", ".md"]

# RoadRunner internal container configuration (docs: https://github.com/spiral/endure).
endure:
  # How long to wait for stopping.
  #
  # Default: 30s
  grace_period: 30s

  # Print graph in the graphviz format to the stdout (paste here to visualize https://dreampuf.github.io)
  #
  # Default: false
  print_graph: false

  # Logging level. Possible values: "debug", "info", "warning", "error", "panic", "fatal".
  #
  # Default: "error"
  log_level: error

```

## Minimal configuration
You are not required to enable every plugin to make RoadRunner work. Given configuration only enables essential plugins
to make HTTP endpoints work:

```yaml
rpc:
  listen: tcp://127.0.0.1:6001

server:
  command: "php tests/psr-worker-bench.php"

http:
  address: "0.0.0.0:8080"
  pool:
    num_workers: 4   
```

## Console flags

You can overwrite any of the config values using `-o` flag:

```
rr serve -o http.address=:80 -o http.workers.pool.numWorkers=1
```

> The values will be merged with `.rr` file.

## Environment Variables

RoadRunner will replace some config options using reference(s) to environment variable(s):

```yaml
http:
  pool.num_workers: ${NUM_WORKERS}
```
