<?php

use Spiral\Goridge;
use Spiral\RoadRunner;

ini_set('display_errors', 'stderr');
require dirname(__DIR__) . "/vendor/autoload.php";

if (count($argv) < 3) {
    die("need 2 arguments");
}

list($test, $goridge) = [$argv[1], $argv[2]];

switch ($goridge) {
    case "pipes":
        $relay = new Goridge\StreamRelay(STDIN, STDOUT);
        break;

    case "tcp":
        $relay = new Goridge\SocketRelay("localhost", 9007);
        break;

    case "unix":
        $relay = new Goridge\SocketRelay(
            "sock.unix",
            null,
            Goridge\SocketRelay::SOCK_UNIX
        );
        break;

    default:
        die("invalid protocol selection");
}

$psr7 = new RoadRunner\Http\PSR7Worker(
    new RoadRunner\Worker($relay),
    new \Nyholm\Psr7\Factory\Psr17Factory(),
    new \Nyholm\Psr7\Factory\Psr17Factory(),
    new \Nyholm\Psr7\Factory\Psr17Factory()
);

require_once sprintf("%s/%s.php", __DIR__, $test);

while ($req = $psr7->waitRequest()) {
    try {
        $psr7->respond(handleRequest($req, new \Nyholm\Psr7\Response()));
    } catch (\Throwable $e) {
        $psr7->getWorker()->error((string)$e);
    }
}
