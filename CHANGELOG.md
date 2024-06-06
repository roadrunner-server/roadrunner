# CHANGELOG

# 🚀 v2024.1.3 🚀


---

# 🚀 v2024.1.2 🚀

### Upcoming new JOBS driver: `google-pub-sub`:
- 🔥: Currently in a polishing phase, the new plugin will be released as part of the `v2024.2.0`. Stay tuned! (thanks @cv65kr)

### `gRPC` plugin:
- 🐛: strip extra slashes when there is no package defined in the protofile: [PR](https://github.com/roadrunner-server/grpc/pull/134), (thanks @satdeveloping)

### `OTEL` plugin:
- 🐛: Fix hardcoded AlwaysSample samples: [BUG](https://github.com/roadrunner-server/roadrunner/issues/1918), (thanks @bazilmarkov)

### `RR core` plugin:
- 🐛: RR `workers/reset` commands don't respect default config values: [BUG](https://github.com/roadrunner-server/roadrunner/issues/1914), (thanks @r4m-alexd)

---

# 🚀 v2024.1.1 🚀

### `HTTP` plugin:
- 🐛 Bug: Fix for the NPE on types check: [BUG](https://github.com/roadrunner-server/roadrunner/issues/1903), (thanks @cto-asocial)

### `gRPC` plugin:
- 🔥 Remove experimental status from the OTEL in `gRPC`, [PR](https://github.com/roadrunner-server/grpc/pull/133)

### `SDK`:
- 🔥 Additional debug logging for the `maxExecs` with `jitter`: [PR](https://github.com/roadrunner-server/sdk/pull/121) (thanks @Kaspiman)

---

# 🚀 v2024.1.0 🚀

## Upgrade guide: [link](https://docs.roadrunner.dev/general/compatibility)

### `HTTP` plugin:
- 🔥 Use `protobuf` encoded payloads to prevent field reordering and JSON escaped symbols.

### `Kafka` driver:
- 🔥 Support [TLS configuration](https://docs.roadrunner.dev/queues-and-jobs/kafka#configuration) (thanks @dkomarek)

### `SDK`:
- 🔥 Use a small random jitter to prevent the [Thundering herd problem](https://en.wikipedia.org/wiki/Thundering_herd_problem) when user uses `max_jobs` option and all the workers restarted at the same time. This feature is enabled automatically. (thanks @Kaspiman)
