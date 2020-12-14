<?php

/**
 * High-performance PHP process supervisor and load balancer written in Go
 *
 * @author Alex Bond
 */
declare(strict_types=1);

namespace Spiral\RoadRunner\Http;

use Spiral\RoadRunner\WorkerInterface;

class HttpWorker
{
    /** @var WorkerInterface */
    private WorkerInterface $worker;

    /**
     * @param WorkerInterface $worker
     */
    public function __construct(WorkerInterface $worker)
    {
        $this->worker = $worker;
    }

    /**
     * @return WorkerInterface
     */
    public function getWorker(): WorkerInterface
    {
        return $this->worker;
    }

    /**
     * Wait for incoming http request.
     *
     * @return Request|null
     */
    public function waitRequest(): ?Request
    {
        $payload = $this->getWorker()->waitPayload();
        if (empty($payload->body) && empty($payload->header)) {
            // termination request
            return null;
        }

        $request = new Request();
        $request->body = $payload->body;

        $context = json_decode($payload->header, true);
        if ($context === null) {
            // invalid context
            return null;
        }

        $this->hydrateRequest($request, $context);

        return $request;
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
        if ($headers === []) {
            // this is required to represent empty header set as map and not as array
            $headers = new \stdClass();
        }

        $this->getWorker()->send(
            $body,
            (string) json_encode(['status' => $status, 'headers' => $headers])
        );
    }

    /**
     * @param Request $request
     * @param array   $context
     */
    private function hydrateRequest(Request $request, array $context): void
    {
        $request->remoteAddr = $context['remoteAddr'];
        $request->protocol = $context['protocol'];
        $request->method = $context['method'];
        $request->uri = $context['uri'];
        $request->attributes = $context['attributes'];
        $request->headers = $context['headers'];
        $request->cookies = $context['cookies'];
        $request->uploads = $context['uploads'];
        parse_str($context['rawQuery'], $request->query);

        // indicates that body was parsed
        $request->parsed = $context['parsed'];
    }
}
