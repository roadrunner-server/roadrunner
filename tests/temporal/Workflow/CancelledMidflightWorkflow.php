<?php

namespace Temporal\Tests\Workflow;

use Temporal\Activity\ActivityOptions;
use Temporal\Tests\Activity\SimpleActivity;
use Temporal\Workflow;
use Temporal\Workflow\WorkflowMethod;

#[Workflow\WorkflowInterface]
class CancelledMidflightWorkflow
{
    private array $status = [];

    #[Workflow\QueryMethod(name: 'getStatus')]
    public function getStatus(): array
    {
        return $this->status;
    }

    #[WorkflowMethod(name: 'CancelledMidflightWorkflow')]
    public function handler()
    {
        $simple = Workflow::newActivityStub(
            SimpleActivity::class,
            ActivityOptions::new()->withStartToCloseTimeout(5)
        );

        $this->status[] = 'start';

        $scope = Workflow::newCancellationScope(
            function () use ($simple) {
                $this->status[] = 'in scope';
                $simple->slow('1');
            }
        )->onCancel(
            function () {
                $this->status[] = 'on cancel';
            }
        );

        $scope->cancel();
        $this->status[] = 'done cancel';

        return 'OK';
    }
}
