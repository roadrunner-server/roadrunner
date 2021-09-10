# Server Commands

RoadRunner application can be started by calling a simple command from the root of your PHP application.

```
$ rr serve 
```

You can also start RoadRunner using configuration from custom location:

```
$ rr serve -c ./app/.rr.yaml
```

To reload all RoadRunner services:

```
$ rr reset
```

> You can attach this command as file watcher in your IDE.

To reset only particular plugins:

```
$ rr reset http
```

To run golang pprof server (debug mode):

```
$ rr serve -d -c .rr.yaml
```

To view the status of all active workers in an interactive mode.

```
$ rr workers -i
```

```
Workers of [http]:
+---------+-----------+---------+---------+-----------------+
|   PID   |  STATUS   |  EXECS  | MEMORY  |     CREATED     |
+---------+-----------+---------+---------+-----------------+
|    9440 | ready     |  42,320 | 31 MB   | 22 days ago     |
|    9447 | ready     |  42,329 | 31 MB   | 22 days ago     |
|    9454 | ready     |  42,306 | 31 MB   | 22 days ago     |
|    9461 | ready     |  42,316 | 31 MB   | 22 days ago     |
+---------+-----------+---------+---------+-----------------+
```
