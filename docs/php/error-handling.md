# Error Handling
There are multiple ways of how you can handle errors produces by PHP workers.

The simplest and most common way would be responding to parent service with the error message using `getWorker()->error()`:

```php
try {
    $resp = new \Zend\Diactoros\Response();
    $resp->getBody()->write("hello world");

    $psr7->respond($resp);
} catch (\Throwable $e) {
    $psr7->getWorker()->error((string)$e);
}
```

You can also flush your warning and errors into `STDERR` to output them directly into the console (similar to docker-compose).

```php
file_put_contents('php://stderr', 'my message');
```

Since RoadRunner 2.0 all warnings send to STDOUT will be forwarded to STDERR as well.