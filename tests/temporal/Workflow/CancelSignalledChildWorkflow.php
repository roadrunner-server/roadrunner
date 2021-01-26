<?php

namespace Temporal\Tests\Workflow;

use React\Promise\Deferred;
use Temporal\Workflow;
use Temporal\Workflow\WorkflowMethod;

#[Workflow\WorkflowInterface]
class CancelSignalledChildWorkflow
{
    private array $status = [];

    #[Workflow\QueryMethod(name: 'getStatus')]
    public function getStatus(): array
    {
        return $this->status;
    }

    #[WorkflowMethod(name: 'CancelSignalledChildWorkflow')]
    public function handler()
    {
        // typed stub
        $simple = Workflow::newChildWorkflowStub(SimpleSignalledWorkflow::class);

        $waitSignalled = new Deferred();

        $this->status[] = 'start';

        // start execution
        $scope = Workflow::newCancellationScope(
            function () use ($simple, $waitSignalled) {
                $call = $simple->handler();
                $this->status[] = 'child started';

                yield $simple->add(8);
                $this->status[] = 'child signalled';
                $waitSignalled->resolve();

                return yield $call;
            }
        );

        // only cancel scope when signal dispatched
        yield $waitSignalled;
        $scope->cancel();
        $this->status[] = 'scope cancelled';

        try {
            return yield $scope;
        } catch (\Throwable $e) {
            $this->status[] = 'process done';

            return 'cancelled ok';
        }
    }
}
