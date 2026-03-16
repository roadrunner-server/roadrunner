<?php
/**
 * Sample GRPC PHP server.
 */

use Spiral\RoadRunner\GRPC\ContextInterface;
use Health\HealthInterface;
use Health\HealthCheckRequest;
use Health\HealthCheckResponse;

class HealthService implements HealthInterface
{
    public function Check(ContextInterface $ctx, HealthCheckRequest $in): HealthCheckResponse
    {
        $out = new HealthCheckResponse();
        $out->setStatus(HealthCheckResponse\ServingStatus::SERVING);
        return $out;
    }

    public function Watch(ContextInterface $ctx, HealthCheckRequest $in): HealthCheckResponse
    {
        $out = new HealthCheckResponse();
        $out->setStatus(HealthCheckResponse\ServingStatus::SERVING);
        return $out;
    }
}
