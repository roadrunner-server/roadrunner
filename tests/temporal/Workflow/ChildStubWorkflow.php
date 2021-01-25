<?php

namespace Temporal\Tests\Workflow;

use Temporal\Workflow;
use Temporal\Workflow\WorkflowMethod;

#[Workflow\WorkflowInterface]
class ChildStubWorkflow
{
    #[WorkflowMethod(name: 'ChildStubWorkflow')]
    public function handler(
        string $input
    ) {
        // typed stub
        $simple = Workflow::newChildWorkflowStub(SimpleWorkflow::class);

        $result = [];
        $result[] = yield $simple->handler($input);

        // untyped
        $untyped = Workflow::newUntypedChildWorkflowStub('SimpleWorkflow');
        $result[] = yield $untyped->execute(['untyped']);

        $execution = yield $untyped->getExecution();
        assert($execution instanceof Workflow\WorkflowExecution);

        return $result;
    }
}
