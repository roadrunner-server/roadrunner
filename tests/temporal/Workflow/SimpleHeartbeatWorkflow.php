<?php

declare(strict_types=1);

namespace Temporal\Tests\Workflow;

use Temporal\Activity\ActivityOptions;
use Temporal\Workflow;
use Temporal\Workflow\WorkflowMethod;
use Temporal\Tests\Activity\HeartBeatActivity;

#[Workflow\WorkflowInterface]
class SimpleHeartbeatWorkflow
{
    #[WorkflowMethod(name: 'SimpleHeartbeatWorkflow')]
    public function handler(int $iterations): iterable
    {
        $act = Workflow::newActivityStub(
            HeartBeatActivity::class,
            ActivityOptions::new()->withStartToCloseTimeout(50)
        );

        return yield $act->doSomething($iterations);
    }
}
