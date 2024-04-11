# CHANGELOG

# 🚀 v2024.1.0 🚀

## Upgrade guide: [link](https://docs.roadrunner.dev/general/compatibility)

### `HTTP` plugin:
- 🔥 Use `protobuf` encoded payloads to prevent field reordering and JSON escaped symbols.

### `Kafka` driver:
- 🔥 Support [TLS configuration](https://docs.roadrunner.dev/queues-and-jobs/kafka#configuration) (thanks @dkomarek)

### `SDK`:
- 🔥 Use a small random jitter to prevent the [Thundering herd problem](https://en.wikipedia.org/wiki/Thundering_herd_problem) when user uses `max_jobs` option and all the workers restarted at the same time. This feature is enabled automatically. (thanks @Kaspiman)
