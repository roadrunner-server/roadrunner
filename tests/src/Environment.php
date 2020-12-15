<?php

/**
 * High-performance PHP process supervisor and load balancer written in Go.
 *
 * @author Wolfy-J
 */

declare(strict_types=1);

namespace Spiral\RoadRunner;

use Spiral\RoadRunner\Exception\EnvironmentException;

class Environment implements EnvironmentInterface
{
    /** @var array */
    private array $env;

    /**
     * @param array $env
     */
    public function __construct(array $env)
    {
        $this->env = $env;
    }

    /**
     * Returns worker mode assigned to the PHP process.
     *
     * @return string
     * @throws EnvironmentException
     */
    public function getMode(): string
    {
        return $this->getValue('RR_MODE');
    }

    /**
     * Address worker should be connected to (or pipes).
     *
     * @return string
     * @throws EnvironmentException
     */
    public function getRelayAddress(): string
    {
        return $this->getValue('RR_RELAY');
    }

    /**
     * RPC address.
     *
     * @return string
     * @throws EnvironmentException
     */
    public function getRPCAddress(): string
    {
        return $this->getValue('RR_RPC');
    }

    /**
     * @param string $name
     * @return string
     * @throws EnvironmentException
     */
    private function getValue(string $name): string
    {
        if (!isset($this->env[$name])) {
            throw new EnvironmentException(sprintf("Missing environment value `%s`", $name));
        }

        return (string) $this->env[$name];
    }

    /**
     * @return EnvironmentInterface
     */
    public static function fromGlobals(): EnvironmentInterface
    {
        return new static(array_merge($_SERVER, $_ENV));
    }
}
