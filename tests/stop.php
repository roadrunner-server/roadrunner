<?php
/**
 * @var Goridge\RelayInterface $relay
 */

use Spiral\Goridge;
use Spiral\RoadRunner;

$rr = new RoadRunner\Worker($relay);

$used = false;
while ($in = $rr->waitPayload()) {
    try {
        if ($used) {
            // kill on second attempt
            $rr->stop();
            continue;
        }

        $used = true;
        $rr->respond(new RoadRunner\Payload((string)getmypid()));
    } catch (\Throwable $e) {
        $rr->error((string)$e);
    }
}
