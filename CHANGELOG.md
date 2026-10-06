# CHANGELOG

releases: [docs](https://docs.roadrunner.dev/docs/releases)

## v3.0.0

- Move the Go module and import paths to `github.com/roadrunner-server/roadrunner/v3`.
- Update CI and release builds for v3 version metadata and container image tags.
- Add Zstd response compression and the Protoreg descriptor registry to the default plugins.
- Add the bundled `rate_limiter` HTTP middleware with global, IP, and header keys.
