<?php

use Spiral\Goridge;
use Spiral\RoadRunner;

/**
 * echo client over pipes.
 */
ini_set('display_errors', 'stderr');
require "vendor/autoload.php";

$rr = new RoadRunner\Worker(new Goridge\StreamRelay(STDIN, STDOUT));

while ($in = $rr->receive($ctx)) {
    try {
        $rr->send((string)$in);
    } catch (\Throwable $e) {
        $rr->error((string)$e);
    }
}