<?php
 /**
  * @var Goridge\RelayInterface $relay
  */

 use Spiral\Goridge;
 use Spiral\RoadRunner;

 $rr = new RoadRunner\Worker($relay);

 while ($in = $rr->receive($ctx)) {
     try {
         sleep(1);
         $rr->send((string)getmypid());
     } catch (\Throwable $e) {
         $rr->error((string)$e);
     }
 }