<?php
/**
 * @var Goridge\RelayInterface $relay
 */

use Spiral\Goridge;
use Spiral\RoadRunner;

$rr = new RoadRunner\Worker($relay);

while ($in = $rr->waitPayload()) {
    try {
        usleep($in->body * 1000);
        $rr->respond(new RoadRunner\Payload(''));
    } catch (\Throwable $e) {
        $rr->error((string)$e);
    }
}
