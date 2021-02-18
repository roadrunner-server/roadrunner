<?php
/**
 * @var Goridge\RelayInterface $relay
 */

use Spiral\Goridge;
use Spiral\RoadRunner;

require dirname(__DIR__) . "/../vendor/autoload.php";

$relay = new Goridge\SocketRelay(
            "unix.sock",
            null,
            Goridge\SocketRelay::SOCK_UNIX
        );

$rr = new RoadRunner\Worker($relay);

while ($in = $rr->waitPayload()) {
    try {
        $rr->respond(new RoadRunner\Payload((string)$in->body));
    } catch (\Throwable $e) {
        $rr->error((string)$e);
    }
}
