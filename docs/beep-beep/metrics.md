# Application Metrics

RoadRunner server includes an embedded metrics server based on [Prometheus](https://prometheus.io/).

## Enable Metrics

To enable metrics add `metrics` section to your configuration:

```yaml
metrics:
  address: localhost:2112
```

Once complete you can access Prometheus metrics using `http://localhost:2112/metrics` url.

Make sure to install metrics extension:

```bash
composer require spiral/roadrunner-metrics
```

## Application metrics

You can also publish application-specific metrics using an RPC connection to the server. First, you have to register a metric in your
configuration file:

```yaml
metrics:
  address: localhost:2112
  collect:
    app_metric_counter:
      type: counter
      help: "Application counter."
```

To send metric from the application:

```php
$metrics = new RoadRunner\Metrics\Metrics(
    Goridge\RPC\RPC::create(RoadRunner\Environment::fromGlobals()->getRPCAddress())
);

$metrics->add('app_metric_counter', 1);
```

> Supported types: gauge, counter, summary, histogram.

## Tagged metrics

You can use tagged (labels) metrics to group values:

```yaml
metrics:
  address: localhost:2112
  collect:
    app_type_duration:
      type: histogram
      help: "Application counter."
      labels: ["type"]
```

You should specify values for your labels while pushing the metric:

```php
$metrics = new RoadRunner\Metrics\Metrics(
    Goridge\RPC\RPC::create(RoadRunner\Environment::fromGlobals()->getRPCAddress())
);

$metrics->add('app_type_duration', 0.5, ['some-type']);
```

## Declare metrics

You can declare metric from PHP application itself:

```php
$metrics->declare(
    'test',
    RoadRunner\Metrics\Collector::counter()->withHelp('Test counter')
);
```
