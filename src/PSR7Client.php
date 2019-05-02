<?php
/**
 * High-performance PHP process supervisor and load balancer written in Go
 *
 * @author Wolfy-J
 */
declare(strict_types=1);

namespace Spiral\RoadRunner;

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
    /** @var HttpClient */
    private $httpClient;

    /** @var ServerRequestFactoryInterface */
    private $requestFactory;

    /** @var StreamFactoryInterface */
    private $streamFactory;

    /*** @var UploadedFileFactoryInterface */
    private $uploadsFactory;

    private $originalServer = [];

    /** @var array Valid values for HTTP protocol version */
    private static $allowedVersions = ['1.0', '1.1', '2',];

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
        $this->httpClient = new HttpClient($worker);
        $this->requestFactory = $requestFactory ?? new Diactoros\ServerRequestFactory();
        $this->streamFactory = $streamFactory ?? new Diactoros\StreamFactory();
        $this->uploadsFactory = $uploadsFactory ?? new Diactoros\UploadedFileFactory();
        $this->originalServer = $_SERVER;
    }

    /**
     * @return Worker
     */
    public function getWorker(): Worker
    {
        return $this->httpClient->getWorker();
    }

    /**
     * @return ServerRequestInterface|null
     */
    public function acceptRequest()
    {
        $rawRequest = $this->httpClient->acceptRequest();
        if ($rawRequest === null) {
            return null;
        }

        $_SERVER = $this->configureServer($rawRequest['ctx']);

        $request = $this->requestFactory->createServerRequest(
            $rawRequest['ctx']['method'],
            $rawRequest['ctx']['uri'],
            $_SERVER
        );

        parse_str($rawRequest['ctx']['rawQuery'], $query);

        $request = $request
            ->withProtocolVersion(static::fetchProtocolVersion($rawRequest['ctx']['protocol']))
            ->withCookieParams($rawRequest['ctx']['cookies'])
            ->withQueryParams($query)
            ->withUploadedFiles($this->wrapUploads($rawRequest['ctx']['uploads']));

        foreach ($rawRequest['ctx']['attributes'] as $name => $value) {
            $request = $request->withAttribute($name, $value);
        }

        foreach ($rawRequest['ctx']['headers'] as $name => $value) {
            $request = $request->withHeader($name, $value);
        }

        if ($rawRequest['ctx']['parsed']) {
            $request = $request->withParsedBody(json_decode($rawRequest['body'], true));
        } else {
            if ($rawRequest['body'] !== null) {
                $request = $request->withBody($this->streamFactory->createStream($rawRequest['body']));
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
        $this->httpClient->respond(
            $response->getStatusCode(),
            $response->getBody()->__toString(),
            $response->getHeaders()
        );
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
        $server = $this->originalServer;
        $server['REQUEST_TIME'] = time();
        $server['REQUEST_TIME_FLOAT'] = microtime(true);
        $server['REMOTE_ADDR'] = $ctx['attributes']['ipAddress'] ?? $ctx['remoteAddr'] ?? '127.0.0.1';

        $server['HTTP_USER_AGENT'] = '';
        foreach ($ctx['headers'] as $key => $value) {
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
