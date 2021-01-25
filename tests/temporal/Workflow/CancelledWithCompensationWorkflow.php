<?php

namespace Temporal\Tests\Workflow;

use Temporal\Activity\ActivityOptions;
use Temporal\Exception\Failure\CanceledFailure;
use Temporal\Tests\Activity\SimpleActivity;
use Temporal\Workflow;
use Temporal\Workflow\WorkflowMethod;

#[Workflow\WorkflowInterface]
class CancelledWithCompensationWorkflow
{
    private array $status = [];

    #[Workflow\QueryMethod(name: 'getStatus')]
    public function getStatus(): array
    {
        return $this->status;
    }

    #[WorkflowMethod(name: 'CancelledWithCompensationWorkflow')]
    public function handler()
    {
        $simple = Workflow::newActivityStub(
            SimpleActivity::class,
            ActivityOptions::new()->withStartToCloseTimeout(5)
        );

        // waits for 2 seconds
        $slow = $simple->slow('DOING SLOW ACTIVITY');

        try {
            $this->status[] = 'yield';
            $result = yield $slow;
        } catch (CanceledFailure $e) {
            $this->status[] = 'rollback';

            try {
                // must fail again
                $result = yield $slow;
            } catch (CanceledFailure $e) {
                $this->status[] = 'captured retry';
            }

            try {
                // fail since on cancelled context
                $result = yield $simple->echo('echo must fail');
            } catch (CanceledFailure $e) {
                $this->status[] = 'captured promise on cancelled';
            }

            $scope = Workflow::newDetachedCancellationScope(
                function () use ($simple) {
                    $this->status[] = 'START rollback';

                    $second = yield $simple->echo('rollback');

                    $this->status[] = sprintf("RESULT (%s)", $second);

                    if ($second !== 'ROLLBACK') {
                        $this->status[] = 'FAIL rollback';
                        return 'failed to compensate ' . $second;
                    }
                    $this->status[] = 'DONE rollback';

                    return 'OK';
                }
            );

            $this->status[] = 'WAIT ROLLBACK';
            $result = yield $scope;
            $this->status[] = 'COMPLETE rollback';
        }

        $this->status[] = 'result: ' . $result;
        return $result;
    }
}
