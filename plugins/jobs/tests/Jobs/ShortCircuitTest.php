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
use Spiral\Jobs\Options;
use Spiral\Jobs\Registry\ContainerRegistry;
use Spiral\Jobs\ShortCircuit;
use Spiral\Jobs\Tests\Local\ErrorJob;
use Spiral\Jobs\Tests\Local\Job;

class ShortCircuitTest extends TestCase
{
    protected function tearDown(): void
    {
        if (file_exists(Job::JOB_FILE)) {
            unlink(Job::JOB_FILE);
        }
    }

    public function testLocal(): void
    {
        $c = new ContainerRegistry(new Container());
        $jobs = new ShortCircuit($c, $c);

        $id = $jobs->push(Job::class, ['data' => 100]);

        $this->assertNotEmpty($id);

        $this->assertFileExists(Job::JOB_FILE);

        $data = json_decode(file_get_contents(Job::JOB_FILE), true);
        $this->assertSame($id, $data['id']);
        $this->assertSame(100, $data['data']);
    }

    public function testLocalDelayed(): void
    {
        $c = new ContainerRegistry(new Container());
        $jobs = new ShortCircuit($c, $c);

        $t = microtime(true);
        $id = $jobs->push(Job::class, ['data' => 100], Options::delayed(1));

        $this->assertTrue(microtime(true) - $t >= 1);

        $this->assertNotEmpty($id);

        $this->assertFileExists(Job::JOB_FILE);

        $data = json_decode(file_get_contents(Job::JOB_FILE), true);
        $this->assertSame($id, $data['id']);
        $this->assertSame(100, $data['data']);
    }

    /**
     * @expectedException \Spiral\Jobs\Exception\JobException
     */
    public function testError(): void
    {
        $c = new ContainerRegistry(new Container());
        $jobs = new ShortCircuit($c, $c);
        $jobs->push(ErrorJob::class);
    }

    public function testLocalDelay(): void
    {
        $c = new ContainerRegistry(new Container());
        $jobs = new ShortCircuit($c, $c);

        $id = $jobs->push(Job::class, ['data' => 100], Options::delayed(1));
        $this->assertNotEmpty($id);

        $this->assertFileExists(Job::JOB_FILE);

        $data = json_decode(file_get_contents(Job::JOB_FILE), true);
        $this->assertSame($id, $data['id']);
        $this->assertSame(100, $data['data']);
    }
}
