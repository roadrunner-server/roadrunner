<?php

declare(strict_types=1);

namespace Temporal\Tests\Workflow;

use Temporal\Activity\ActivityOptions;
use Temporal\Common\RetryOptions;
use Temporal\Workflow;
use Temporal\Workflow\WorkflowMethod;
use Temporal\Tests\Activity\HeartBeatActivity;

#[Workflow\WorkflowInterface]
class FailedHeartbeatWorkflow
{
    #[WorkflowMethod(name: 'FailedHeartbeatWorkflow')]
    public function handler(
        int $iterations
    ): iterable {
        $act = Workflow::newActivityStub(
            HeartBeatActivity::class,
            ActivityOptions::new()
                ->withStartToCloseTimeout(50)
                // will fail on first attempt
                ->withRetryOptions(RetryOptions::new()->withMaximumAttempts(2))
        );

        return yield $act->failedActivity($iterations);
    }
}
