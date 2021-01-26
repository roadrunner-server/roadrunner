<?php

namespace Temporal\Tests\Workflow;

use Temporal\Activity\ActivityOptions;
use Temporal\Tests\Activity\SimpleActivity;
use Temporal\Workflow;
use Temporal\Workflow\WorkflowMethod;

#[Workflow\WorkflowInterface]
class ActivityStubWorkflow
{
    #[WorkflowMethod(name: 'ActivityStubWorkflow')]
    public function handler(
        string $input
    ) {
        // typed stub
        $simple = Workflow::newActivityStub(
            SimpleActivity::class,
            ActivityOptions::new()->withStartToCloseTimeout(5)
        );

        $result = [];
        $result[] = yield $simple->echo($input);

        try {
            $simple->undefined($input);
        } catch (\BadMethodCallException $e) {
            $result[] = 'invalid method call';
        }

        // untyped stub
        $untyped = Workflow::newUntypedActivityStub(ActivityOptions::new()->withStartToCloseTimeout(1));

        $result[] = yield $untyped->execute('SimpleActivity.echo', ['untyped']);

        return $result;
    }
}
