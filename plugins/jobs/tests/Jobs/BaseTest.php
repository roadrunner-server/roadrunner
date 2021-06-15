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
use Spiral\Goridge\RPC;
use Spiral\Goridge\SocketRelay;
use Spiral\Jobs\Options;
use Spiral\Jobs\Queue;
use Spiral\Jobs\Registry\ContainerRegistry;

abstract class BaseTest extends TestCase
{
    public const JOB       = null;
    public const ERROR_JOB = null;

    private $job;
    private $errorJob;

    public function setUp(): void
    {
        $this->job = static::JOB;
        $this->errorJob = static::ERROR_JOB;
    }

    protected function tearDown(): void
    {
        if (file_exists((static::JOB)::JOB_FILE)) {
            unlink((static::JOB)::JOB_FILE);
        }
    }

    public function testJob(): void
    {
        $jobs = $this->makeJobs();

        $id = $jobs->push($this->job, ['data' => 100]);

        $this->assertNotEmpty($id);

        $this->waitForJob();
        $this->assertFileExists($this->job::JOB_FILE);

        $data = json_decode(file_get_contents($this->job::JOB_FILE), true);
        $this->assertSame($id, $data['id']);
        $this->assertSame(100, $data['data']);
    }

    public function testErrorJob(): void
    {
        $jobs = $this->makeJobs();

        $id = $jobs->push($this->errorJob, ['data' => 100]);
        $this->assertNotEmpty($id);
    }

    public function testDelayJob(): void
    {
        $jobs = $this->makeJobs();

        $id = $jobs->push($this->job, ['data' => 100], Options::delayed(1));

        $this->assertNotEmpty($id);

        $this->assertTrue($this->waitForJob() > 1);
        $this->assertFileExists($this->job::JOB_FILE);

        $data = json_decode(file_get_contents($this->job::JOB_FILE), true);
        $this->assertSame($id, $data['id']);
        $this->assertSame(100, $data['data']);
    }

    /**
     * @expectedException \Spiral\Jobs\Exception\JobException
     */
    public function testConnectionException(): void
    {
        $jobs = new Queue(
            new RPC(new SocketRelay('localhost', 6002)),
            new ContainerRegistry(new Container())
        );

        $jobs->push($this->job, ['data' => 100]);
    }

    public function makeJobs(): Queue
    {
        return new Queue(
            new RPC(new SocketRelay('localhost', 6001)),
            new ContainerRegistry(new Container())
        );
    }

    private function waitForJob(): float
    {
        $start = microtime(true);
        $try = 0;
        while (!file_exists($this->job::JOB_FILE) && $try < 10) {
            usleep(250000);
            $try++;
        }

        return microtime(true) - $start;
    }
}
