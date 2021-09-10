# Developer Mode
RoadRunner uses PHP scripts in daemon mode, this means that you have to reload a server every time you change your codebase. 

If you use any modern IDE you can achieve that by adding file watcher which automatically invokes command `rr reset` for the plugins specified in the `reload` config.

> Or use [auto-resetter](/beep-beep/reload.md).

## In Docker
You can reset rr process in docker by connecting to it using local rr client. 

```yaml
rpc:
  listen: tcp://:6001
```

> Make sure to forward/expose port 6001.

Then run `rr reset` locally on file change.

## Debug Mode
To run workers in debug mode (similar to how PHP-FPM operates):

```yaml
http:
  pool.debug: true
```
