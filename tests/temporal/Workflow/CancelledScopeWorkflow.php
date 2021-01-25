<?php

namespace Temporal\Tests\Workflow;

use Temporal\Activity\ActivityOptions;
use Temporal\Workflow;
use Temporal\Workflow\WorkflowMethod;
use Temporal\Tests\Activity\SimpleActivity;

#[Workflow\WorkflowInterface]
class CancelledScopeWorkflow
{
    #[WorkflowMethod(name: 'CancelledScopeWorkflow')]
    public function handler()
    {
        $simple = Workflow::newActivityStub(
            SimpleActivity::class,
            ActivityOptions::new()->withStartToCloseTimeout(5)
        );

        $cancelled = 'not';

        $scope = Workflow::newCancellationScope(
            function () use ($simple) {
                yield Workflow::timer(2);
                yield $simple->slow('hello');
            }
        )->onCancel(
            function () use (&$cancelled) {
                $cancelled = 'yes';
            }
        );

        yield Workflow::timer(1);
        $scope->cancel();

        return $cancelled;
    }
}
