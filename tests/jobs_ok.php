<?php

/**
 * @var Goridge\RelayInterface $relay
 */

use Spiral\Goridge;
use Spiral\RoadRunner;
use Spiral\Goridge\StreamRelay;

require __DIR__ . "/vendor/autoload.php";

$rr = new RoadRunner\Worker(new StreamRelay(\STDIN, \STDOUT));

while ($in = $rr->waitPayload()) {
    try {
        $ctx = json_decode($in->header, true);
        $headers = $ctx['headers'];

        $rr->respond(new RoadRunner\Payload(json_encode([
            'type' => 0,
            'data' => []
        ])));
    } catch (\Throwable $e) {
        $rr->error((string)$e);
    }
}
