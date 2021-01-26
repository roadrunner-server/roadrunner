<?php

namespace Temporal\Tests\Workflow;

use Temporal\Exception\Failure\CanceledFailure;
use Temporal\Workflow;
use Temporal\Workflow\WorkflowMethod;

#[Workflow\WorkflowInterface]
class CancelledNestedWorkflow
{
    private array $status = [];

    #[Workflow\QueryMethod(name: 'getStatus')]
    public function getStatus(): array
    {
        return $this->status;
    }

    #[WorkflowMethod(name: 'CancelledNestedWorkflow')]
    public function handler()
    {
        $this->status[] = 'begin';
        try {
            yield Workflow::newCancellationScope(
                function () {
                    $this->status[] = 'first scope';

                    $scope = Workflow::newCancellationScope(
                        function () {
                            $this->status[] = 'second scope';

                            try {
                                yield Workflow::timer(2);
                            } catch (CanceledFailure $e) {
                                $this->status[] = 'second scope cancelled';
                                throw $e;
                            }

                            $this->status[] = 'second scope done';
                        }
                    )->onCancel(
                        function () {
                            $this->status[] = 'close second scope';
                        }
                    );

                    try {
                        yield Workflow::timer(1);
                    } catch (CanceledFailure $e) {
                        $this->status[] = 'first scope cancelled';
                        throw $e;
                    }

                    $this->status[] = 'first scope done';

                    yield $scope;
                }
            )->onCancel(
                function () {
                    $this->status[] = 'close first scope';
                }
            );
        } catch (CanceledFailure $e) {
            $this->status[] = 'close process';

            return 'CANCELLED';
        }

        return 'OK';
    }
}
