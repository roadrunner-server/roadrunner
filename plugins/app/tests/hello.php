<?php
/**
 * @var Goridge\RelayInterface $relay
 */

use Spiral\Goridge;
use Spiral\RoadRunner;

require dirname(__DIR__) . "/../../vendor_php/autoload.php";

$relay = new Goridge\StreamRelay(STDIN, STDOUT);
$rr = new RoadRunner\Worker($relay);

while ($in = $rr->receive($ctx)) {
    try {
        $rr->send((string)$in);
    } catch (\Throwable $e) {
        $rr->error((string)$e);
    }
}