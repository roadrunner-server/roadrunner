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
     * @return mixed[]|null Request information as ['ctx'=>[], 'body'=>string]
     *                      or null if termination request or invalid context.
     */
    public function acceptRequest(): ?array
    {
        $body = $this->getWorker()->receive($ctx);
        if (empty($body) && empty($ctx)) {
            // termination request
            return null;
        }

        $ctx = json_decode($ctx, true);
        if ($ctx === null) {
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
    public function respond(int $status, string $body, array $headers = []): void
    {
        $sendHeaders = empty($headers)
            ? new \stdClass() // this is required to represent empty header set as map and not as array
            : $headers;

        $this->getWorker()->send(
            $body,
            (string) json_encode(['status' => $status, 'headers' => $sendHeaders])
        );
    }
}
