<?php
/**
 * @var Goridge\RelayInterface $relay
 */

use Spiral\Goridge;
use Spiral\RoadRunner;

$rr = new RoadRunner\Worker($relay);

while ($in = $rr->receive($ctx)) {
    echo undefined_function();
    $rr->send((string)$in);
}