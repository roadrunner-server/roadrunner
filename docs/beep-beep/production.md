# Production Usage

There are multiple tips and suggestions which must be acknowledged while running RoadRunner on production.

## State and memory

State and memory are **not shared** between different worker instances but are **shared** for a single worker instance.
Since a single worker typically process more than a single request, you should be careful about it:

- Make sure to close all descriptors (especially in case of fatal exceptions).
- [optional] consider calling `gc_collect_cycles` after each execution if you want to keep the memory low
  (this will slow down your application a bit).
- Watch memory leaks - you have to be more picky about what components you use. Workers will be restarted in case of
  a memory leak, but it should not be hard to completely avoid this issue by properly designing your application.
- Avoid state pollution (i.e. globals or user data cache in memory).
- Database connections and any pipe/socket is the potential point of failure. Simple way of dealing with it is to close
  all connections after each iteration. Note that it is not the most performant solution.
  
## Configuration

- Make sure NOT to listen 0.0.0.0 in RPC service (unless in Docker).
- Connect to a worker using pipes for higher performance (Unix sockets just a bit slower).
- Tweak your pool timings to the values you like.
- A number of workers = number of CPU threads in your system, unless your application is IO bound, then pick
  the number heuristically. 
- Consider using `max_jobs` for your workers if you experience any issues with application stability over time.
- RoadRunner is +40% performant using Keep-Alive connections.
- Set memory limit to least 10-20% below `max_memory_usage`.
- Since RoadRunner workers run from cli you need to enable OPcache in CLI via `opcache.enable_cli=1`.
- Make sure to use [health check endpoint](beep-beep/health.md) when running rr in a cloud environment.
- Use `user` option in the config to start workers processes from the particular user on Linux based systems.
