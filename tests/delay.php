<?php
/**
 * @var Goridge\RelayInterface $relay
 */

use Spiral\Goridge;
use Spiral\RoadRunner;

$rr = new RoadRunner\Worker($relay);

while ($in = $rr->receive($ctx)) {
    try {
        usleep($in * 1000);
        $rr->send('');
    } catch (\Throwable $e) {
        $rr->error((string)$e);
    }
}