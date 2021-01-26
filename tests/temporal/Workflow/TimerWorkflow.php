<?php

declare(strict_types=1);

namespace Temporal\Tests\Workflow;

use Temporal\Activity\ActivityOptions;
use Temporal\Workflow;
use Temporal\Workflow\WorkflowMethod;
use Temporal\Tests\Activity\SimpleActivity;

#[Workflow\WorkflowInterface]
class TimerWorkflow
{
    #[WorkflowMethod(name: 'TimerWorkflow')]
    public function handler(string $input): iterable
    {
        $simple = Workflow::newActivityStub(
            SimpleActivity::class,
            ActivityOptions::new()->withStartToCloseTimeout(5)
        );

        yield Workflow::timer(1);

        return yield $simple->lower($input);
    }
}
