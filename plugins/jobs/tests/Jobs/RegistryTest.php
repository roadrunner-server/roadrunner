<?php

/**
 * Spiral Framework.
 *
 * @license   MIT
 * @author    Anton Titov (Wolfy-J)
 */

declare(strict_types=1);

namespace Spiral\Jobs\Tests;

use PHPUnit\Framework\TestCase;
use Spiral\Core\Container;
use Spiral\Jobs\Registry\ContainerRegistry;
use Spiral\Jobs\Tests\Local\Job;

class RegistryTest extends TestCase
{
    public function testMakeJob(): void
    {
        $factory = new ContainerRegistry(new Container());

        $j = $factory->getHandler('spiral.jobs.tests.local.job');
        $this->assertInstanceOf(Job::class, $j);

        $this->assertSame(json_encode(['data' => 200]), $j->serialize(
            'spiral.jobs.tests.local.job',
            ['data' => 200]
        ));
    }

    /**
     * @expectedException \Spiral\Jobs\Exception\JobException
     */
    public function testMakeUndefined(): void
    {
        $factory = new ContainerRegistry(new Container());

        $factory->getHandler('spiral.jobs.undefined');
    }
}
