<?php

declare(strict_types=1);

use Spiral\Goridge\StreamRelay;
use Spiral\RoadRunner\Worker as RoadRunner;

require dirname(__DIR__) . "/vendor_php/autoload.php";

$rr = new RoadRunner(new StreamRelay(\STDIN, \STDOUT));
$mem = '';
while($rr->receive($ctx)){
    $mem .= str_repeat(" ", 1024*1024);
    $rr->send("");
}