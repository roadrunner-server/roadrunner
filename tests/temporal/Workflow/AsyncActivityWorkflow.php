<?php

declare(strict_types=1);

namespace Temporal\Tests\Workflow;

use Temporal\Activity\ActivityCancellationType;
use Temporal\Activity\ActivityOptions;
use Temporal\Workflow;
use Temporal\Workflow\WorkflowMethod;
use Temporal\Tests\Activity\SimpleActivity;

#[Workflow\WorkflowInterface]
class AsyncActivityWorkflow
{
    #[WorkflowMethod(name: 'AsyncActivityWorkflow')]
    public function handler()
    {
        $simple = Workflow::newActivityStub(
            SimpleActivity::class,
            ActivityOptions::new()
                ->withStartToCloseTimeout(20)
                ->withCancellationType(ActivityCancellationType::WAIT_CANCELLATION_COMPLETED)
        );

        return yield $simple->external();
    }
}
