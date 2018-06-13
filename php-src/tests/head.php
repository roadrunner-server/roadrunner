<?php
/**
 * @var Goridge\RelayInterface $relay
 */

use Spiral\Goridge;
use Spiral\RoadRunner;

$rr = new RoadRunner\Worker($relay);

while ($in = $rr->receive($ctx)) {
    try {
        $rr->send("", (string)$ctx);
    } catch (\Throwable $e) {
        $rr->error((string)$e);
    }
}