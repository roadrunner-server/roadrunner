<?php

namespace Temporal\Tests\Workflow;

use Temporal\Activity\ActivityOptions;
use Temporal\Common\RetryOptions;
use Temporal\Tests\Activity\SimpleActivity;
use Temporal\Workflow;
use Temporal\Workflow\WorkflowMethod;

#[Workflow\WorkflowInterface]
class ComplexExceptionalWorkflow
{
    #[WorkflowMethod(name: 'ComplexExceptionalWorkflow')]
    public function handler()
    {
        $child = Workflow::newChildWorkflowStub(
            ExceptionalActivityWorkflow::class,
            Workflow\ChildWorkflowOptions::new()->withRetryOptions(
                (new RetryOptions())->withMaximumAttempts(1)
            )
        );

        return yield $child->handler();
    }
}
