<?php


namespace Temporal\Tests\Workflow;


use Temporal\Workflow;
use Temporal\Workflow\SignalMethod;
use Temporal\Workflow\WorkflowInterface;
use Temporal\Workflow\WorkflowMethod;

#[WorkflowInterface]
class WaitWorkflow
{
    private bool $ready = false;
    private string $value;

    #[SignalMethod]
    public function unlock(
        string $value
    ) {
        $this->ready = true;
        $this->value = $value;
    }

    #[WorkflowMethod(name: 'WaitWorkflow')]
    public function run()
    {
        yield Workflow::await(fn() => $this->ready);

        return $this->value;
    }
}
