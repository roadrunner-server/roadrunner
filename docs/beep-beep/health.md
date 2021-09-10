# Health Endpoint
RoadRunner server includes a health check endpoint that returns the health of the workers.

## Enable health

To enable the health check endpoint, add a `status` section to your configuration:

```yaml
status:
  address: localhost:2114
```

To access the health-check use the following URL:

`http://localhost:2114/health?plugin=http`

> You can check one or multiple plugins using health-check. Currently, only HTTP supported.

Once enabled, the health check endpoint will respond with the following: 

 - `HTTP 200` if there is at least **one worker** ready to serve requests.
 - `HTTP 500` if there are **no workers** ready to service requests.

## Use cases

The health check endpoint can be used for the following:

 - [Kubernetes readiness and liveness probes](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/)
 - [AWS target group health checks](https://docs.aws.amazon.com/elasticloadbalancing/latest/application/target-group-health-checks.html)
 - [GCE Load Balancing health checks](https://cloud.google.com/load-balancing/docs/health-checks)
