<?php

/**
 * High-performance PHP process supervisor and load balancer written in Go
 *
 * @author Wolfy-J
 */
declare(strict_types=1);

namespace Spiral\RoadRunner\Diactoros;

use Psr\Http\Message\ServerRequestFactoryInterface;
use Psr\Http\Message\ServerRequestInterface;
use Laminas\Diactoros\ServerRequest;

final class ServerRequestFactory implements ServerRequestFactoryInterface
{
    /**
     * @inheritdoc
     *
     * @param array<mixed> $serverParams Array of SAPI parameters with which to seed the generated request instance.
     */
    public function createServerRequest(string $method, $uri, array $serverParams = []): ServerRequestInterface
    {
        $uploadedFiles = [];
        return new ServerRequest($serverParams, $uploadedFiles, $uri, $method);
    }
}
