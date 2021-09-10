# Serving static content

It is possible to serve static content using RoadRunner.

## Enable HTTP Middleware

To enable static content serving use the configuration inside the http section:

```yaml
http:
  # host and port separated by semicolon
  address: 127.0.0.1:44933
  # ...
    static:
      dir: "."
      forbid: [""]
      allow: [".txt", ".php"]
      calculate_etag: false
      weak: false
      request:
        input: "custom-header"
      response:
        output: "output-header"
```

Where:

1. `dir`: path to the directory.
3. `forbid`: file extensions that should not be served.
4. `allow`: file extensions which should be served (empty - serve all except forbidden). If extension presented in both (allow and forbid) hashmaps - that treated as we should forbid file extension.
5. `calculate_etag`: turn on etag computation for the static file.
6. `weak`: use a weak generator (/W), it uses only filename to generate a CRC32 sum. If false - all file content used to generate CRC32 sum.
7. `request/response`: custom headers for the static files.  

To combine static content with other middleware, use the following sequence (static will always be the last in the row, file server will apply headers and gzip plugins):

```yaml
http:
  # host and port separated by semicolon
  address: 127.0.0.1:44933
  # ...
  middleware: [ "headers", "gzip" ]
  # ...
  headers:
    # ...
    static:
      dir: "."
      forbid: [""]
      allow: [".txt", ".php"]
      calculate_etag: false
      weak: false
      request:
        input: "custom-header"
      response:
        output: "output-header"
```
