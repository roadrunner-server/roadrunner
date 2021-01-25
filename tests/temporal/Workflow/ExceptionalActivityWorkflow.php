<?php

namespace Temporal\Tests\Workflow;

use Temporal\Activity\ActivityOptions;
use Temporal\Common\RetryOptions;
use Temporal\Tests\Activity\SimpleActivity;
use Temporal\Workflow;
use Temporal\Workflow\WorkflowMethod;

#[Workflow\WorkflowInterface]
class ExceptionalActivityWorkflow
{
    #[WorkflowMethod(name: 'ExceptionalActivityWorkflow')]
    public function handler()
    {
        $simple = Workflow::newActivityStub(
            SimpleActivity::class,
            ActivityOptions::new()->withStartToCloseTimeout(5)
                ->withRetryOptions((new RetryOptions())->withMaximumAttempts(1))
        );

        return yield $simple->fail();
    }
}
