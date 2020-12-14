<?php

/**
 * High-performance PHP process supervisor and load balancer written in Go.
 *
 * @author Wolfy-J
 */

declare(strict_types=1);

namespace Spiral\RoadRunner;

use Spiral\Goridge\Exception\GoridgeException;
use Spiral\RoadRunner\Exception\RoadRunnerException;

interface WorkerInterface
{
    /**
     * Wait for incoming payload from the server. Must return null when worker stopped.
     *
     * @return Payload|null
     * @throws GoridgeException
     * @throws RoadRunnerException
     */
    public function waitPayload(): ?Payload;

    /**
     * Respond to the server with the processing result.
     *
     * @param Payload $payload
     * @throws GoridgeException
     */
    public function respond(Payload $payload): void;

    /**
     * Respond to the server with an error. Error must be treated as TaskError and might not cause
     * worker destruction.
     *
     * Example:
     *
     * $worker->error("invalid payload");
     *
     * @param string $error
     * @throws GoridgeException
     */
    public function error(string $error): void;

    /**
     * Terminate the process. Server must automatically pass task to the next available process.
     * Worker will receive stop command after calling this method.
     *
     * Attention, you MUST use continue; after invoking this method to let rr to properly stop worker.
     */
    public function stop(): void;
}
