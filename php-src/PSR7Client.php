<?php
/**
 * High-performance PHP process supervisor and load balancer written in Go
 *
 * @author Wolfy-J
 */

namespace Spiral\RoadRunner;

use Psr\Http\Message\ResponseInterface;
use Psr\Http\Message\ServerRequestInterface;
use Spiral\RoadRunner\Worker;
use Zend\Diactoros\ServerRequest;
use Zend\Diactoros\Stream;
use Zend\Diactoros\UploadedFile;

/**
 * Spiral Framework, SpiralScout LLC.
 *
 * @package   spiralFramework
 * @author    Anton Titov (Wolfy-J)
 * @copyright ©2009-2011
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
            $bodyStream = new Stream("php://memory", "rwb");
            $bodyStream->write($body);
        }

        return new ServerRequest(
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
    }

    /**
     * Send response to the application server.
     *
     * @param ResponseInterface $response
     */
    public function respond(ResponseInterface $response)
    {
        $this->worker->send($response->getBody(), json_encode([
            'status'  => $response->getStatusCode(),
            'headers' => $response->getHeaders()
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
            if (isset($file['name'])) {
                $result[$index] = new UploadedFile(
                    $file['tmpName'],
                    $file['size'],
                    $file['error'],
                    $file['name'],
                    $file['type']
                );
                continue;
            }

            $result[$index] = $this->wrapUploads($file);
        }

        return $result;
    }
}