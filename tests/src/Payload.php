<?php

/**
 * High-performance PHP process supervisor and load balancer written in Go.
 *
 * @author Wolfy-J
 */

declare(strict_types=1);

namespace Spiral\RoadRunner;

/**
 * Class Payload
 *
 * @package Spiral\RoadRunner
 */
final class Payload
{
    /**
     * Execution payload (binary).
     *
     * @var string|null
     */
    public ?string $body;

    /**
     * Execution context (binary).
     *
     * @var string|null
     */
    public ?string $header;

    /**
     * @param string|null $body
     * @param string|null $header
     */
    public function __construct(?string $body, ?string $header = null)
    {
        $this->body = $body;
        $this->header = $header;
    }
}
