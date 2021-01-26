<?php


namespace Temporal\Tests\Workflow;

use Temporal\Activity\ActivityOptions;
use Temporal\Workflow;
use Temporal\Workflow\WorkflowMethod;
use Temporal\Tests\Activity\SimpleActivity;

#[Workflow\WorkflowInterface]
class WorkflowWithSequence
{
    #[WorkflowMethod(name: 'WorkflowWithSequence')]
    public function handler()
    {
        $simple = Workflow::newActivityStub(
            SimpleActivity::class,
            ActivityOptions::new()->withStartToCloseTimeout(5)
        );

        $a = $simple->echo('a');
        $b = $simple->echo('b');

        yield $a;
        yield $b;

        return 'OK';
    }
}
