<?php
 /**
  * @var Goridge\RelayInterface $relay
  */

 use Spiral\Goridge;
 use Spiral\RoadRunner;

 $rr = new RoadRunner\Worker($relay);

 while ($in = $rr->waitPayload()) {
     try {
         sleep(1);
         $rr->respond(new RoadRunner\Payload((string)getmypid()));
     } catch (\Throwable $e) {
         $rr->error((string)$e);
     }
 }
