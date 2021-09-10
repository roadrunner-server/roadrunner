# Logging

RoadRunner provides the ability to control the log for each plugin individually.

## Global configuration

To configure logging globally use `logs` config section:

```yaml
logs:
  mode: production
  output: stderr
```

To use develop mode. It enables development mode (which makes DPanicLevel logs panic), uses a console encoder, writes to
standard error, and disables sampling. Stacktraces are automatically included on logs of WarnLevel and above.

```yaml
logs:
  mode: development
```

To output to separate file:

```yaml
logs:
  mode: production
  output: file.log
```

To use console friendly output:

```yaml
logs:
  encoding: console # default value
```

To suppress messages under specific log level:

```yaml
logs:
  encoding: console # default value
  level: info
```

### File logger  

It is possibe to redirect channels or whole log output to the file:

Config sample with file logger:

```yaml
logs:
  mode: development
  level: debug
  file_logger_options:
    log_output: "test.log"
    max_size: 10
    max_age: 24
    max_backups: 10
    compress: true
```

OR for the log channel:

```yaml
logs:
  mode: development
  level: debug
  channels:
      http:
          file_logger_options:
              log_output: "test.log"
              max_size: 10
              max_age: 24
              max_backups: 10
              compress: true
```

1. `log_output`: Filename is the file to write logs to in the same directory.  It uses <processname>-lumberjack.log in os.TempDir() if empty.
2. `max_size`: is the maximum size in megabytes of the log file before it gets rotated. It defaults to 100 megabytes.
3. `max_age`: is the maximum number of days to retain old log files based on the timestamp encoded in their filename.  Note that a day is defined as 24 hours and may not exactly correspond to calendar days due to daylight savings, leap seconds, etc. The default is not to remove old log files based on age.
4. `max_backups`: is the maximum number of old log files to retain.  The default is to retain all old log files (though MaxAge may still cause them to get deleted.)
5. `compress`: determines if the rotated log files should be compressed using gzip. The default is not to perform compression.

## Channels

In addition, you can configure each plugin log messages individually using `channels` section:

```yaml
logs:
  encoding: console # default value
  level: info
  channels:
    server.mode: none # disable server logging. Also `off` can be used.
    http:
      mode: production
      output: http.log
```

## Summary

1. Levels: `panic`, `error`, `warn`, `info`, `debug`. Default: `debug`.
2. Encodings: `console`, `json`. Default: `console`.
3. Modes: `production`, `development`, `raw`. Default: `development`.
4. Output: `file.log` or `stderr`, `stdout`. Default `stderr`.
5. Error output: `err_file.log` or `stderr`, `stdout`. Default `stderr`.

> Feel free to register your own [ZapLogger](https://github.com/uber-go/zap) extensions.
