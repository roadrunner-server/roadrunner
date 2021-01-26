<?php

namespace Temporal\Tests\Workflow;

use Temporal\Workflow;
use Temporal\Workflow\SignalMethod;
use Temporal\Workflow\WorkflowInterface;
use Temporal\Workflow\WorkflowMethod;

#[WorkflowInterface]
class AggregatedWorkflow
{
    private array $values = [];

    #[SignalMethod]
    public function addValue(
        string $value
    ) {
        $this->values[] = $value;
    }

    #[WorkflowMethod(name: 'AggregatedWorkflow')]
    public function run(
        int $count
    ) {
        yield Workflow::await(fn() => count($this->values) === $count);

        return $this->values;
    }
}
