# Debugging
You can use RoadRunner scripts with xDebug extension. In order to enable configure your IDE to accept remote connections. 

Note, if you run multiple PHP processes you have to extend the maximum number of allowed connections to the number of 
active workers, otherwise some calls would not be caught on your breakpoints.

![xdebug](https://user-images.githubusercontent.com/796136/46493729-c767b400-c819-11e8-9110-505a256994b0.png)

To activate xDebug make sure to set the `xdebug.mode=debug` in your `php.ini`. 

To enable xDebug in your application make sure to set ENV variable `XDEBUG_SESSION`:

```
rpc:
   listen: tcp://127.0.0.1:6001

server:
   command: "php worker.php"
   env:
      XDEBUG_SESSION: 1

http:
   address: "0.0.0.0:8080"
   pool:
      num_workers: 1
      debug: true
```

You should be able to use breakpoints and view state at this point.
