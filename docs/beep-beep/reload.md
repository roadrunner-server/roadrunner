# Auto-Reloading
RoadRunner is able to automatically detect PHP file changes and reload connected services. Such approach allows you to develop application without the `max_jobs: 1` or manual server reset.

## Configuration
To enable reloading for http service:

```yaml
reload:
  # sync interval
  interval: 1s
  # global patterns to sync
  patterns: [ ".php" ]
  # list of included for sync services
  services:
    http:
      # recursive search for file patterns to add
      recursive: true
      # ignored folders
      ignore: [ "vendor" ]
      # service specific file pattens to sync
      patterns: [ ".php", ".go", ".md" ]
      # directories to sync. If recursive is set to true,
      # recursive sync will be applied only to the directories in `dirs` section
      dirs: [ "." ]
```

## Performance
The `reload` component will affect the performance of application server. Make sure to use it in development mode only. In the future we have plans to rewrite this plugin to use native OS capabilities of notification events.
