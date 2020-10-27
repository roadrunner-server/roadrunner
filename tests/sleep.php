<?php

declare(strict_types=1);

use Spiral\Goridge\StreamRelay;
use Spiral\RoadRunner\Worker as RoadRunner;

require dirname(__DIR__) . "/vendor_php/autoload.php";

$rr = new RoadRunner(new StreamRelay(\STDIN, \STDOUT));

while($rr->receive($ctx)){
    sleep(3);
    $rr->send("");
}