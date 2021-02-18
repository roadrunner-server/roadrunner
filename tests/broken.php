<?php
/**
 * @var Goridge\RelayInterface $relay
 */

use Spiral\Goridge;
use Spiral\RoadRunner;

$rr = new RoadRunner\Worker($relay);

while ($in = $rr->waitPayload()) {
    echo undefined_function();
    $rr->respond(new RoadRunner\Payload((string)$in->body, null));
}
