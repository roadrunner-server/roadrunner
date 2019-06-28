<?php
/**
 * Spiral Framework.
 *
 * @license   MIT
 * @author    Anton Titov (Wolfy-J)
 */
declare(strict_types=1);

namespace Spiral\RoadRunner;

use Spiral\Goridge\Exceptions\RPCException;
use Spiral\Goridge\RPC;
use Spiral\RoadRunner\Exception\MetricException;

/**
 * Application metrics.
 */
final class Metrics implements MetricsInterface
{
    /** @var RPC */
    private $rpc;

    /**
     * @param RPC $rpc
     */
    public function __construct(RPC $rpc)
    {
        $this->rpc = $rpc;
    }

    /**
     * Add collector value. Fallback to appropriate method of related collector.
     *
     * @param string $collector
     * @param float  $value
     * @param array  $labels
     *
     * @throws MetricException
     */
    public function add(string $collector, float $value, array $labels = [])
    {
        try {
            $this->rpc->call('metrics.Add', compact('collector', 'value', 'labels'));
        } catch (RPCException $e) {
            throw new MetricException($e->getMessage(), $e->getCode(), $e);
        }
    }

    /**
     * Subtract the collector value, only for gauge collector.
     *
     * @param string $collector
     * @param float  $value
     * @param array  $labels
     *
     * @throws MetricException
     */
    public function sub(string $collector, float $value, array $labels = [])
    {
        try {
            $this->rpc->call('metrics.Sub', compact('collector', 'value', 'labels'));
        } catch (RPCException $e) {
            throw new MetricException($e->getMessage(), $e->getCode(), $e);
        }
    }

    /**
     * Observe collector value, only for histogram and summary collectors.
     *
     * @param string $collector
     * @param float  $value
     * @param array  $labels
     *
     * @throws MetricException
     */
    public function observe(string $collector, float $value, array $labels = [])
    {
        try {
            $this->rpc->call('metrics.Observe', compact('collector', 'value', 'labels'));
        } catch (RPCException $e) {
            throw new MetricException($e->getMessage(), $e->getCode(), $e);
        }
    }

    /**
     * Set collector value, only for gauge collector.
     *
     * @param string $collector
     * @param float  $value
     * @param array  $labels
     *
     * @throws MetricException
     */
    public function set(string $collector, float $value, array $labels = [])
    {
        try {
            $this->rpc->call('metrics.Set', compact('collector', 'value', 'labels'));
        } catch (RPCException $e) {
            throw new MetricException($e->getMessage(), $e->getCode(), $e);
        }
    }
}