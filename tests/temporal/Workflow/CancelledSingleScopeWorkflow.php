<?php

namespace Temporal\Tests\Workflow;

use Temporal\Activity\ActivityOptions;
use Temporal\Exception\Failure\CanceledFailure;
use Temporal\Workflow;
use Temporal\Workflow\WorkflowMethod;
use Temporal\Tests\Activity\SimpleActivity;

#[Workflow\WorkflowInterface]
class CancelledSingleScopeWorkflow
{
    private array $status = [];

    #[Workflow\QueryMethod(name: 'getStatus')]
    public function getStatus(): array
    {
        return $this->status;
    }

    #[WorkflowMethod(name: 'CancelledSingleScopeWorkflow')]
    public function handler()
    {
        $simple = Workflow::newActivityStub(
            SimpleActivity::class,
            ActivityOptions::new()
                ->withStartToCloseTimeout(5)
        );

        $this->status[] = 'start';
        try {
            yield Workflow::newCancellationScope(
                function () use ($simple) {
                    try {
                        $this->status[] = 'in scope';
                        yield $simple->slow('1');
                    } catch (CanceledFailure $e) {
                        // after process is complete, do not use for business logic
                        $this->status[] = 'captured in scope';
                        throw $e;
                    }
                }
            )->onCancel(
                function () {
                    $this->status[] = 'on cancel';
                }
            );
        } catch (CanceledFailure $e) {
            $this->status[] = 'captured in process';
        }

        return 'OK';
    }
}
