<?php
/**
 * High-performance PHP process supervisor and load balancer written in Go
 *
 * @author Alex Bond
 */
declare(strict_types=1);

namespace Spiral\RoadRunner;

final class HttpClient
{
    /** @var Worker */
    private $worker;

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
}
