<?php

declare(strict_types=1);

use Spiral\Goridge\StreamRelay;
use Spiral\RoadRunner\Worker as RoadRunner;

require __DIR__ . "/vendor/autoload.php";

if (file_exists('break')) {
	throw new Exception('oops');
}

$rr = new RoadRunner(new StreamRelay(\STDIN, \STDOUT));

while($rr->waitPayload()){
    $rr->respond(new \Spiral\RoadRunner\Payload(""));
}
