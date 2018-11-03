<?php
/**
 * High-performance PHP process supervisor and load balancer written in Go
 *
 * @author Wolfy-J
 */

namespace Spiral\RoadRunner;

use Http\Factory\Diactoros;
use Psr\Http\Message\ResponseInterface;
use Psr\Http\Message\ServerRequestFactoryInterface;
use Psr\Http\Message\ServerRequestInterface;
use Psr\Http\Message\StreamFactoryInterface;
use Psr\Http\Message\UploadedFileFactoryInterface;

/**
 * Manages PSR-7 request and response.
 */
class PSR7Client
{
    /** @var Worker */
    private $worker;

    /** @var ServerRequestFactoryInterface */
    private $requestFactory;

    /** @var StreamFactoryInterface */
    private $streamFactory;

    /*** @var UploadedFileFactoryInterface */
    private $uploadsFactory;

    /**
     * @param Worker                             $worker
     * @param ServerRequestFactoryInterface|null $requestFactory
     * @param StreamFactoryInterface|null        $streamFactory
     * @param UploadedFileFactoryInterface|null  $uploadsFactory
     */
    public function __construct(
        Worker $worker,
        ServerRequestFactoryInterface $requestFactory = null,
        StreamFactoryInterface $streamFactory = null,
        UploadedFileFactoryInterface $uploadsFactory = null
    ) {
        $this->worker = $worker;
        $this->requestFactory = $requestFactory ?? new Diactoros\ServerRequestFactory();
        $this->streamFactory = $streamFactory ?? new Diactoros\StreamFactory();
        $this->uploadsFactory = $uploadsFactory ?? new Diactoros\UploadedFileFactory();
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

        $_SERVER = [];
        $_SERVER = $this->configureServer($ctx);

        $request = $this->requestFactory->createServerRequest(
            $ctx['method'],
            $ctx['uri'],
            $_SERVER
        );

        parse_str($ctx['rawQuery'], $query);

        $request = $request
            ->withProtocolVersion(substr($ctx['protocol'], 5))
            ->withCookieParams($ctx['cookies'])
            ->withQueryParams($query)
            ->withUploadedFiles($this->wrapUploads($ctx['uploads']));

        foreach ($ctx['attributes'] as $name => $value) {
            $request = $request->withAttribute($name, $value);
        }

        foreach ($ctx['headers'] as $name => $value) {
            $request = $request->withHeader($name, $value);
        }

        if ($ctx['parsed']) {
            $request = $request->withParsedBody(json_decode($body, true));
        } else {
            if ($body !== null) {
                $request = $request->withBody($this->streamFactory->createStream($body));
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
     * Returns altered copy of _SERVER variable. Sets ip-address,
     * request-time and other values.
     *
     * @param array $ctx
     * @return array
     */
    protected function configureServer(array $ctx): array
    {
        $server = $_SERVER;
        $server['REQUEST_TIME'] = time();
        $server['REQUEST_TIME_FLOAT'] = microtime(true);
        $server['REMOTE_ADDR'] = $ctx['attributes']['ipAddress'] ?? $ctx['remoteAddr'] ?? '127.0.0.1';

        return $server;
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
        foreach ($files as $index => $f) {
            if (!isset($f['name'])) {
                $result[$index] = $this->wrapUploads($f);
                continue;
            }

            if (UPLOAD_ERR_OK === $f['error']) {
                $stream = $this->streamFactory->createStreamFromFile($f['tmpName']);
            } else {
                $stream = $this->streamFactory->createStream();
            }

            $result[$index] = $this->uploadsFactory->createUploadedFile(
                $stream,
                $f['size'],
                $f['error'],
                $f['name'],
                $f['mime']
            );
        }

        return $result;
    }
}
