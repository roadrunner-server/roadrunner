<?php

/**
 * Simple jobs consumer that acknowledges every task immediately.
 */

use Spiral\RoadRunner\Jobs\Consumer;

ini_set('display_errors', 'stderr');
require dirname(__DIR__) . "/vendor/autoload.php";

$consumer = new Consumer();

while ($task = $consumer->waitTask()) {
    try {
        $task->complete();
    } catch (\Throwable $e) {
        $task->error((string)$e);
    }
}
