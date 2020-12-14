<?php

/**
 * High-performance PHP process supervisor and load balancer written in Go
 *
 * @author Wolfy-J
 */
declare(strict_types=1);

namespace Spiral\RoadRunner\Http;

use Psr\Http\Message\ResponseInterface;
use Psr\Http\Message\ServerRequestFactoryInterface;
use Psr\Http\Message\ServerRequestInterface;
use Psr\Http\Message\StreamFactoryInterface;
use Psr\Http\Message\UploadedFileFactoryInterface;
use Psr\Http\Message\UploadedFileInterface;
use Spiral\RoadRunner\WorkerInterface;

/**
 * Manages PSR-7 request and response.
 */
class PSR7Worker
{
    private HttpWorker                    $httpWorker;
    private ServerRequestFactoryInterface $requestFactory;
    private StreamFactoryInterface        $streamFactory;
    private UploadedFileFactoryInterface  $uploadsFactory;

    /** @var mixed[] */
    private array $originalServer = [];

    /** @var string[] Valid values for HTTP protocol version */
    private static array $allowedVersions = ['1.0', '1.1', '2',];

    /**
     * @param WorkerInterface               $worker
     * @param ServerRequestFactoryInterface $requestFactory
     * @param StreamFactoryInterface        $streamFactory
     * @param UploadedFileFactoryInterface  $uploadsFactory
     */
    public function __construct(
        WorkerInterface $worker,
        ServerRequestFactoryInterface $requestFactory,
        StreamFactoryInterface $streamFactory,
        UploadedFileFactoryInterface $uploadsFactory
    ) {
        $this->httpWorker = new HttpWorker($worker);
        $this->requestFactory = $requestFactory;
        $this->streamFactory = $streamFactory;
        $this->uploadsFactory = $uploadsFactory;
        $this->originalServer = $_SERVER;
    }

    /**
     * @return WorkerInterface
     */
    public function getWorker(): WorkerInterface
    {
        return $this->httpWorker->getWorker();
    }

    /**
     * @return ServerRequestInterface|null
     */
    public function waitRequest(): ?ServerRequestInterface
    {
        $httpRequest = $this->httpWorker->waitRequest();
        if ($httpRequest === null) {
            return null;
        }

        $_SERVER = $this->configureServer($httpRequest['ctx']);

        return $this->mapRequest($httpRequest, $_SERVER);
    }

    /**
     * Send response to the application server.
     *
     * @param ResponseInterface $response
     */
    public function respond(ResponseInterface $response): void
    {
        $this->httpWorker->respond(
            $response->getStatusCode(),
            $response->getBody()->__toString(),
            $response->getHeaders()
        );
    }

    /**
     * Returns altered copy of _SERVER variable. Sets ip-address,
     * request-time and other values.
     *
     * @param Request $request
     * @return mixed[]
     */
    protected function configureServer(Request $request): array
    {
        $server = $this->originalServer;

        $server['REQUEST_URI'] = $request->uri;
        $server['REQUEST_TIME'] = time();
        $server['REQUEST_TIME_FLOAT'] = microtime(true);
        $server['REMOTE_ADDR'] = $request->getRemoteAddr();
        $server['REQUEST_METHOD'] = $request->method;

        $server['HTTP_USER_AGENT'] = '';
        foreach ($request->headers as $key => $value) {
            $key = strtoupper(str_replace('-', '_', $key));
            if (\in_array($key, ['CONTENT_TYPE', 'CONTENT_LENGTH'])) {
                $server[$key] = implode(', ', $value);
            } else {
                $server['HTTP_' . $key] = implode(', ', $value);
            }
        }

        return $server;
    }

    /**
     * @param Request $httpRequest
     * @param array   $server
     * @return ServerRequestInterface
     */
    protected function mapRequest(Request $httpRequest, array $server): ServerRequestInterface
    {
        $request = $this->requestFactory->createServerRequest(
            $httpRequest->method,
            $httpRequest->uri,
            $_SERVER
        );

        $request = $request
            ->withProtocolVersion(static::fetchProtocolVersion($httpRequest->protocol))
            ->withCookieParams($httpRequest->cookies)
            ->withQueryParams($httpRequest->query)
            ->withUploadedFiles($this->wrapUploads($httpRequest->uploads));

        foreach ($httpRequest->attributes as $name => $value) {
            $request = $request->withAttribute($name, $value);
        }

        foreach ($httpRequest->headers as $name => $value) {
            $request = $request->withHeader($name, $value);
        }

        if ($httpRequest->parsed) {
            return $request->withParsedBody($httpRequest->getParsedBody());
        }

        if ($httpRequest->body !== null) {
            return $request->withBody($this->streamFactory->createStream($httpRequest->body));
        }

        return $request;
    }

    /**
     * Wraps all uploaded files with UploadedFile.
     *
     * @param array[] $files
     * @return UploadedFileInterface[]|mixed[]
     */
    protected function wrapUploads(array $files): array
    {
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

    /**
     * Normalize HTTP protocol version to valid values
     *
     * @param string $version
     * @return string
     */
    private static function fetchProtocolVersion(string $version): string
    {
        $v = substr($version, 5);

        if ($v === '2.0') {
            return '2';
        }

        // Fallback for values outside of valid protocol versions
        if (!in_array($v, static::$allowedVersions, true)) {
            return '1.1';
        }

        return $v;
    }
}
