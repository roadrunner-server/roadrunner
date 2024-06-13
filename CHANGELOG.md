# CHANGELOG

# 🚀 v2024.1.4 🚀

### `Temporal` plugin:
- 🐛: Fix Local activities executed on the Workflow PHP Worker instead of the Activity PHP Worker: [BUG](https://github.com/roadrunner-server/roadrunner/issues/1940). With this fix, LA performance should see a significant increase. (thanks @Zylius)


---

# 🚀 v2024.1.3 🚀

### `RR core`:
- 🔥: Deprecate `RR_*` env variables prefix. This was an undocumented feature which caused confusion, because any configuration value might be automatically replaced (without using env in the configuration) with a matching `RR_*` environment variable, [PR](https://github.com/roadrunner-server/roadrunner/pull/1929).

### `Temporal` plugin:
- 🔥: Enable client certificate rotation: [FR](https://github.com/temporalio/roadrunner-temporal/issues/522). With this change you may replace certificate on a Live system. (thanks @benkelukas)
- 🔥: Expose `continue_as_new_suggested` for the PHP Worker: [PR](https://github.com/temporalio/roadrunner-temporal/pull/520).

### `Kafka`
- 🐛: Reduce number of `maxPollRecords` from 10k to 100, [PR](https://github.com/roadrunner-server/kafka/commit/f7950cb538e6c670cfc50681e61eb939c591f27b).

### `Endure` container:
- 🐛: Fix incorrectly used error log message: [PR](https://github.com/roadrunner-server/endure/pull/175).

### General:
- 🔥: Update Go to `v1.22.4`.

## RoadRunner PHP:

### `Worker`:
- 🔥: Add `RR_VERSION` env to the `Environment` class: [PR](https://github.com/roadrunner-php/worker/pull/37), (thanks @Kaspiman)


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
