<?php
/**
 * Spiral Framework.
 *
 * @license   MIT
 * @author    Anton Titov (Wolfy-J)
 */
declare(strict_types=1);

namespace Spiral\RoadRunner;

use Spiral\RoadRunner\Exception\MetricException;

interface MetricsInterface
{
    /**
     * Add collector value. Fallback to appropriate method of related collector.
     *
     * @param string $collector
     * @param float  $value
     * @param array  $labels
     *
     * @throws MetricException
     */
    public function add(string $collector, float $value, array $labels = []);

    /**
     * Subtract the collector value, only for gauge collector.
     *
     * @param string $collector
     * @param float  $value
     * @param array  $labels
     *
     * @throws MetricException
     */
    public function sub(string $collector, float $value, array $labels = []);

    /**
     * Observe collector value, only for histogram and summary collectors.
     *
     * @param string $collector
     * @param float  $value
     * @param array  $labels
     *
     * @throws MetricException
     */
    public function observe(string $collector, float $value, array $labels = []);

    /**
     * Set collector value, only for gauge collector.
     *
     * @param string $collector
     * @param float  $value
     * @param array  $labels
     *
     * @throws MetricException
     */
    public function set(string $collector, float $value, array $labels = []);
}