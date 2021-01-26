<?php

declare(strict_types=1);

namespace Temporal\Tests\Workflow;

use Temporal\Workflow;
use Temporal\Workflow\WorkflowMethod;

#[Workflow\WorkflowInterface]
class SimpleSignalledWorkflowWithSleep
{
    private $counter = 0;

    #[Workflow\SignalMethod(name: "add")]
    public function add(
        int $value
    ) {
        $this->counter += $value;
    }

    #[WorkflowMethod(name: 'SimpleSignalledWorkflowWithSleep')]
    public function handler(): iterable
    {
        // collect signals during one second
        yield Workflow::timer(1);

        if (!Workflow::isReplaying()) {
            sleep(1);
        }

        return $this->counter;
    }
}
