<?php
declare(strict_types=1);

/**
 * High-performance PHP process supervisor and load balancer written in Go
 *
 * @author Alex Bond
 */

namespace Spiral\RoadRunner;

use ReflectionFunction;

class HttpClient
{
    /** @var Worker */
    private $worker;

    /** @var []callable Array of termination event listeners */
    private $terminationListeners = [];

    /**
     * @param Worker $worker
     */
    public function __construct(Worker $worker)
    {
        $this->worker = $worker;
    }

    /**
     * @return Worker
     */
    public function getWorker(): Worker
    {
        return $this->worker;
    }

    /**
     * @return array|null Request information as ['ctx'=>[], 'body'=>string] or null if termination request or invalid context.
     */
    public function acceptRequest()
    {
        $body = $this->getWorker()->receive($ctx);
        if (empty($body) && empty($ctx)) {
            // termination request

            foreach ($this->terminationListeners as $listener) {
                call_user_func($listener);
            }

            return null;
        }

        $ctx = json_decode($ctx, true);
        if (is_null($ctx)) {
            // invalid context
            return null;
        }

        return ['ctx' => $ctx, 'body' => $body];
    }

    /**
     * Send response to the application server.
     *
     * @param int        $status  Http status code
     * @param string     $body    Body of response
     * @param string[][] $headers An associative array of the message's headers. Each
     *                            key MUST be a header name, and each value MUST be an array of strings
     *                            for that header.
     */
    public function respond(int $status, string $body, array $headers = [])
    {
        if (empty($headers)) {
            // this is required to represent empty header set as map and not as array
            $headers = new \stdClass();
        }

        $this->getWorker()->send(
            $body,
            json_encode(['status' => $status, 'headers' => $headers])
        );
    }

    /**
     * Adds function to termination event listeners
     *
     * @param callable $func Function to call on termination request
     * @throws \ReflectionException
     */
    public function addTerminationListener(callable $func)
    {
        $f = new ReflectionFunction($func);
        if ($f->getNumberOfParameters() > 0) {
            throw new \InvalidArgumentException('Termination event callback can\'t have parameters.');
        }
        $this->terminationListeners[] = $func;
    }

    /**
     * Clears all termination event listeners
     */
    public function clearTerminationListeners()
    {
        $this->terminationListeners = [];
    }
}
