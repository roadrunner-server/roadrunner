<?php

declare(strict_types=1);

namespace Temporal\Tests\Workflow;

use Temporal\Workflow;
use Temporal\Workflow\WorkflowMethod;

#[Workflow\WorkflowInterface]
class WithChildStubWorkflow
{
    #[WorkflowMethod(name: 'WithChildStubWorkflow')]
    public function handler(string $input): iterable
    {
        $child = Workflow::newChildWorkflowStub(SimpleWorkflow::class);

        return 'Child: ' . (yield $child->handler('child ' . $input));
    }
}
