<?php

/**
 * High-performance PHP process supervisor and load balancer written in Go.
 *
 * @author Wolfy-J
 */

declare(strict_types=1);

namespace Spiral\RoadRunner;

use Spiral\RoadRunner\Exception\EnvironmentException;

/**
 * Provides base values to configure roadrunner worker.
 */
interface EnvironmentInterface
{
    /**
     * Returns worker mode assigned to the PHP process.
     *
     * @return string
     * @throws EnvironmentException
     */
    public function getMode(): string;

    /**
     * Address worker should be connected to (or pipes).
     *
     * @return string
     * @throws EnvironmentException
     */
    public function getRelayAddress(): string;

    /**
     * RPC address.
     *
     * @return string
     * @throws EnvironmentException
     */
    public function getRPCAddress(): string;
}
