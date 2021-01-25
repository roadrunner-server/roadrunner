<?php

declare(strict_types=1);

namespace Temporal\Tests\Workflow;

use Temporal\Workflow;
use Temporal\Workflow\WorkflowMethod;

#[Workflow\WorkflowInterface]
class WithChildWorkflow
{
    #[WorkflowMethod(name: 'WithChildWorkflow')]
    public function handler(
        string $input
    ): iterable {
        $result = yield Workflow::executeChildWorkflow(
            'SimpleWorkflow',
            ['child ' . $input],
            Workflow\ChildWorkflowOptions::new()
        );

        return 'Child: ' . $result;
    }
}
