### Overriding HTTP default error code

```yaml
http:
  # override http error code for the application errors (default 500)
  appErrorCode: 505
  # override http error code for the internal RR errors (default 500)
  internalErrorCode: 505
```

By default, `http.InternalServerError` code is used, but, for the load balancer might be better to use different code [Feature Request](https://github.com/spiral/roadrunner/issues/471).
These 2 options allow overriding default error code (500) for the internal errors such as `ErrNoPoolAttached` and application error from the PHP.