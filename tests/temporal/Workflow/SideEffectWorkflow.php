<?php

namespace Temporal\Tests\Workflow;

use Temporal\Activity\ActivityOptions;
use Temporal\Common\Uuid;
use Temporal\Workflow;
use Temporal\Workflow\WorkflowMethod;
use Temporal\Tests\Activity\SimpleActivity;

#[Workflow\WorkflowInterface]
class SideEffectWorkflow
{
    #[WorkflowMethod(name: 'SideEffectWorkflow')]
    public function handler(string $input): iterable
    {
        $simple = Workflow::newActivityStub(
            SimpleActivity::class,
            ActivityOptions::new()->withStartToCloseTimeout(5)
        );

        $result = yield Workflow::sideEffect(
            function () use ($input) {
                return $input . '-42';
            }
        );

        return yield $simple->lower($result);
    }
}
