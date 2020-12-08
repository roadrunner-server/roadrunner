<?php

/**
 * High-performance PHP process supervisor and load balancer written in Go
 *
 * @author Wolfy-J
 */
declare(strict_types=1);

namespace Spiral\RoadRunner\Diactoros;

use Psr\Http\Message\StreamInterface;
use Psr\Http\Message\UploadedFileFactoryInterface;
use Psr\Http\Message\UploadedFileInterface;
use Laminas\Diactoros\UploadedFile;

final class UploadedFileFactory implements UploadedFileFactoryInterface
{
    /**
     * @inheritdoc
     */
    public function createUploadedFile(
        StreamInterface $stream,
        int $size = null,
        int $error = \UPLOAD_ERR_OK,
        string $clientFilename = null,
        string $clientMediaType = null
    ): UploadedFileInterface {
        if ($size === null) {
            $size = (int) $stream->getSize();
        }

        /** @var resource $stream */
        return new UploadedFile($stream, $size, $error, $clientFilename, $clientMediaType);
    }
}
