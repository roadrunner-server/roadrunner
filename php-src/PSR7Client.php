<?php
/**
 * High-performance PHP process supervisor and load balancer written in Go
 *
 * @author Wolfy-J
 */

namespace Spiral\RoadRunner;

use Psr\Http\Message\ResponseInterface;
use Psr\Http\Message\ServerRequestInterface;
use Zend\Diactoros;

/**
 * Manages PSR-7 request and response.
 */
class PSR7Client
{
    /**
     * @varWorker
     */
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
     * @return ServerRequestInterface|null
     */
    public function acceptRequest()
    {
        $body = $this->worker->receive($ctx);
        if (empty($body) && empty($ctx)) {
            // termination request
            return null;
        }

        if (empty($ctx = json_decode($ctx, true))) {
            // invalid context
            return null;
        }

        parse_str($ctx['rawQuery'], $query);

        $bodyStream = 'php://input';
        $parsedBody = null;
        if ($ctx['parsed']) {
            $parsedBody = json_decode($body, true);
        } elseif ($body != null) {
            $bodyStream = new Diactoros\Stream("php://memory", "rwb");
            $bodyStream->write($body);
        }

        $request = new Diactoros\ServerRequest(
            $_SERVER,
            $this->wrapUploads($ctx['uploads']),
            $ctx['uri'],
            $ctx['method'],
            $bodyStream,
            $ctx['headers'],
            $ctx['cookies'],
            $query,
            $parsedBody,
            $ctx['protocol']
        );

        if (!empty($ctx['attributes'])) {
            foreach ($ctx['attributes'] as $key => $value) {
                $request = $request->withAttribute($key, $value);
            }
        }

        return $request;
    }

    /**
     * Send response to the application server.
     *
     * @param ResponseInterface $response
     */
    public function respond(ResponseInterface $response)
    {
        $headers = $response->getHeaders();
        if (empty($headers)) {
            // this is required to represent empty header set as map and not as array
            $headers = new \stdClass();
        }

        $this->worker->send($response->getBody(), json_encode([
            'status'  => $response->getStatusCode(),
            'headers' => $headers
        ]));
    }

    /**
     * Wraps all uploaded files with UploadedFile.
     *
     * @param array $files
     *
     * @return array
     */
    private function wrapUploads($files): array
    {
        if (empty($files)) {
            return [];
        }

        $result = [];
        foreach ($files as $index => $file) {
            if (!isset($file['name'])) {
                $result[$index] = $this->wrapUploads($file);
                continue;
            }

            $result[$index] = new Diactoros\UploadedFile(
                $file['tmpName'],
                $file['size'],
                $file['error'],
                $file['name'],
                $file['mime']
            );
        }

        return $result;
    }
}