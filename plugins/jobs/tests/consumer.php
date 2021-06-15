<?php

/**
 * Spiral Framework.
 *
 * @license   MIT
 * @author    Anton Titov (Wolfy-J)
 */

declare(strict_types=1);

use Spiral\Core\Container;
use Spiral\Goridge;
use Spiral\Jobs;
use Spiral\RoadRunner;

require 'bootstrap.php';

$rr = new RoadRunner\Worker(new Goridge\StreamRelay(STDIN, STDOUT));

$consumer = new Jobs\Consumer(new Jobs\Registry\ContainerRegistry(new Container()));
$consumer->serve($rr);
