<?php
/**
 * High-performance PHP process supervisor and load balancer written in Go
 *
 * @author Wolfy-J
 */
declare(strict_types=1);

namespace Spiral\RoadRunner;

use Spiral\Goridge\Exceptions\GoridgeException;
use Spiral\Goridge\RelayInterface as Relay;
use Spiral\RoadRunner\Exception\RoadRunnerException;

/**
 * Accepts connection from RoadRunner server over given Goridge relay.
 *
 * Example:
 *
 * $worker = new Worker(new Goridge\StreamRelay(STDIN, STDOUT));
 * while ($task = $worker->receive($context)) {
 *      $worker->send("DONE", json_encode($context));
 * }
 */
class Worker
{
    // Send as response context to request worker termination
    const STOP = '{"stop":true}';

    /** @var Relay */
    private $relay;

    /**
     * @param Relay $relay
     */
    public function __construct(Relay $relay)
    {
        $this->relay = $relay;
    }

    /**
     * Receive packet of information to process, returns null when process must be stopped. Might
     * return Error to wrap error message from server.
     *
     * @param mixed $header
     * @return \Error|null|string
     *
     * @throws GoridgeException
     */
    public function receive(&$header)
    {
        $body = $this->relay->receiveSync($flags);

        if ($flags & Relay::PAYLOAD_CONTROL) {
            if ($this->handleControl($body, $header, $flags)) {
                // wait for the next command
                return $this->receive($header);
            }

            // no context for the termination.
            $header = null;

            // Expect process termination
            return null;
        }

        if ($flags & Relay::PAYLOAD_ERROR) {
            return new \Error($body);
        }

        return $body;
    }

    /**
     * Respond to the server with result of task execution and execution context.
     *
     * Example:
     * $worker->respond((string)$response->getBody(), json_encode($response->getHeaders()));
     *
     * @param string|null $payload
     * @param string|null $header
     */
    public function send(string $payload = null, string $header = null)
    {
        if (is_null($header)) {
            $this->relay->send($header, Relay::PAYLOAD_CONTROL | Relay::PAYLOAD_NONE);
        } else {
            $this->relay->send($header, Relay::PAYLOAD_CONTROL | Relay::PAYLOAD_RAW);
        }

        $this->relay->send($payload, Relay::PAYLOAD_RAW);
    }

    /**
     * Respond to the server with an error. Error must be treated as TaskError and might not cause
     * worker destruction.
     *
     * Example:
     *
     * $worker->error("invalid payload");
     *
     * @param string $message
     */
    public function error(string $message)
    {
        $this->relay->send(
            $message,
            Relay::PAYLOAD_CONTROL | Relay::PAYLOAD_RAW | Relay::PAYLOAD_ERROR
        );
    }

    /**
     * Terminate the process. Server must automatically pass task to the next available process.
     * Worker will receive StopCommand context after calling this method.
     *
     * Attention, you MUST use continue; after invoking this method to let rr to properly
     * stop worker.
     *
     * @throws GoridgeException
     */
    public function stop()
    {
        $this->send(null, self::STOP);
    }

    /**
     * Handles incoming control command payload and executes it if required.
     *
     * @param string $body
     * @param mixed  $header Exported context (if any).
     * @param int    $flags
     * @return bool True when continue processing.
     *
     * @throws RoadRunnerException
     */
    private function handleControl(string $body = null, &$header = null, int $flags = 0): bool
    {
        $header = $body;
        if (is_null($body) || $flags & Relay::PAYLOAD_RAW) {
            // empty or raw prefix
            return true;
        }

        $p = json_decode($body, true);
        if ($p === false) {
            throw new RoadRunnerException("invalid task context, JSON payload is expected");
        }

        // PID negotiation (socket connections only)
        if (!empty($p['pid'])) {
            $this->relay->send(
                sprintf('{"pid":%s}', getmypid()), Relay::PAYLOAD_CONTROL
            );
        }

        // termination request
        if (!empty($p['stop'])) {
            return false;
        }

        // parsed header
        $header = $p;

        return true;
    }
}