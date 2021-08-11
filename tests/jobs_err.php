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

        $set = isset($headers['attempts']);

        $val = 0;

        if ($set == true) {
            $val = intval($headers['attempts'][0]);
            $val++;
            $headers['attempts'][0] = strval($val);
        } else {
            $headers['attempts'][0] = "1";
        };

        if ($val > 3) {
            $rr->respond(new RoadRunner\Payload(json_encode([
                // no error
                'type' => 0,
                'data' => []
            ])));
        } else {
            $rr->respond(new RoadRunner\Payload(json_encode([
                'type' => 1,
                'data' => [
                    'message' => 'error',
                    'requeue' => true,
                    'delay_seconds' => 5,
                    'headers' => $headers
                ]
            ])));
        }
    } catch (\Throwable $e) {
        $rr->error((string)$e);
    }
}
