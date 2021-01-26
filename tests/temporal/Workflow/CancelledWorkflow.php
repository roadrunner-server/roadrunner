<?php

namespace Temporal\Tests\Workflow;

use Temporal\Activity\ActivityOptions;
use Temporal\Exception\Failure\CanceledFailure;
use Temporal\Tests\Activity\SimpleActivity;
use Temporal\Workflow;
use Temporal\Workflow\WorkflowMethod;

#[Workflow\WorkflowInterface]
class CancelledWorkflow
{
    #[WorkflowMethod(name: 'CancelledWorkflow')]
    public function handler()
    {
        $simple = Workflow::newActivityStub(
            SimpleActivity::class,
            ActivityOptions::new()->withStartToCloseTimeout(5)
        );

        // waits for 2 seconds
        $slow = $simple->slow('DOING SLOW ACTIVITY');

        try {
            return yield $slow;
        } catch (CanceledFailure $e) {
            return "CANCELLED";
        }
    }
}
