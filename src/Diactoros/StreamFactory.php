<?php

/**
 * High-performance PHP process supervisor and load balancer written in Go
 *
 * @author Wolfy-J
 */
declare(strict_types=1);

namespace Spiral\RoadRunner\Diactoros;

use RuntimeException;
use Psr\Http\Message\StreamFactoryInterface;
use Psr\Http\Message\StreamInterface;
use Laminas\Diactoros\Stream;

final class StreamFactory implements StreamFactoryInterface
{
    /**
     * @inheritdoc
     * @throws RuntimeException
     */
    public function createStream(string $content = ''): StreamInterface
    {
        $resource = fopen('php://temp', 'rb+');

        if (! \is_resource($resource)) {
            throw new RuntimeException('Cannot create stream');
        }

        fwrite($resource, $content);
        rewind($resource);
        return $this->createStreamFromResource($resource);
    }

    /**
     * @inheritdoc
     */
    public function createStreamFromFile(string $file, string $mode = 'rb'): StreamInterface
    {
        $resource = fopen($file, $mode);

        if (! \is_resource($resource)) {
            throw new RuntimeException('Cannot create stream');
        }

        return $this->createStreamFromResource($resource);
    }

    /**
     * @inheritdoc
     */
    public function createStreamFromResource($resource): StreamInterface
    {
        return new Stream($resource);
    }
}
