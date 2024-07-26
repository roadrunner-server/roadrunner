# CHANGELOG

# 🚀 v2024.2.0 🚀

# Upgrade guide: [link](https://docs.roadrunner.dev/docs/general/compatibility#upgrading-to-roadrunner-v2024.2.x)

## Community plugins

We are introducing a new term in the RoadRunner community — Community plugins. 
If you have a brilliant idea for the http middleware or JOBS driver or even a new plugin - feel free to check our 
[Customization](../customization) tutorials, create and notify us about your plugin.

## Plugins updates:

### 🔥 Meet the new JOBS driver - Google Pub/Sub
RoadRunner now supports the Google Pub/Sub queues. Currently, this driver is released in **BETA** and has a few limitations which you may find in the [docs]()

### `AMQP` and `Kafka` JOBS drivers

- 🔥 Support an auto-restart pipeline on redial or some fatal problems when connecting to the RabbitMQ broker. Instead of silently exit from the pipeline, RR will try to re-initialize the pipeline.

Thanks to our PHP team, [KV](https://github.com/roadrunner-php/kv/releases/tag/v4.3.0) now has `AsyncStorageInterface` support which makes your experience with the KV plugin even faster.
Feel free to read the technical details here: [link](https://github.com/roadrunner-php/goridge/pull/22)

### Samples repository

- 🔥 Our RoadRunner samples repository was updated and now includes a `Jobs` driver example for the `Jobs` plugin.
More info here: [link](https://github.com/roadrunner-server/samples).


### Our Go-SDK was deprecated

- 😭 Our Go-SDK was deprecated and split into separate packages. Read more in the Upgrade guide.


### Velox configuration update

- 🔥 Velox configuration was simplified:

```yaml
[roadrunner]
# ref -> reference, tag, commit or branch
ref = "v2024.2.0"

# the debug option is used to build RR with debug symbols to profile it with pprof
[debug]
enabled = false

## Rest is the same ....
```

Now, there is no need to include `linker` flags, and buildtime + build version would be inherited automatically.
If you need to debug your binary, please, use the `debug` option set to `true`.

### Special thanks to our sponsors:

1. [Buhta](https://github.com/buhta)
2. [Coderabbitai](https://https://github.com/coderabbitai)
3. [Kaspiman](https://github.com/Kaspiman)
4. [benalf](https://github.com/benalf)
5. [rapita](https://github.com/rapita)
6. [uzulla](https://github.com/uzulla)

---

# 🚀 v2024.1.5 🚀

### `Status` plugin:
- 🐛: Fix k8s-related problem, when status was not available during the graceful shutdown process: [BUG](https://github.com/roadrunner-server/roadrunner/issues/1924). (thanks @cv65kr)

### `JOBS` plugin:
- 🔥: Experimentally added new handlers for the more canonical `ACK`, `NACK`, and `REQUEUE` operations for the `JOBS` drivers. PHP SDK will be updated soon. [FR](https://github.com/roadrunner-server/roadrunner/issues/1941), (thanks @shieldz80)

### ✏️ Future changes:
- 💡: Configuration includes will be out of experimental status in the next minor release (`v2024.2.0`) and currently don't have restrictions on where to put the included config. Keep in mind that the path for the included configurations is calculated from the working directory of the RoadRunner process. [FR](https://github.com/roadrunner-server/roadrunner/issues/935)

---

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
