<?php


namespace Temporal\Tests\Workflow;

use Temporal\Activity\ActivityOptions;
use Temporal\Workflow;
use Temporal\Workflow\WorkflowMethod;
use Temporal\Tests\Activity\SimpleActivity;

#[Workflow\WorkflowInterface]
class ContinuableWorkflow
{
    #[WorkflowMethod(name: 'ContinuableWorkflow')]
    public function handler(
        int $generation
    ) {
        $simple = Workflow::newActivityStub(
            SimpleActivity::class,
            ActivityOptions::new()->withStartToCloseTimeout(5)
        );

        if ($generation > 5) {
            // complete
            return "OK" . $generation;
        }

        if ($generation !== 1) {
            assert(!empty(Workflow::getInfo()->continuedExecutionRunId));
        }

        for ($i = 0; $i < $generation; $i++) {
            yield $simple->echo((string)$generation);
        }

        return Workflow::continueAsNew('ContinuableWorkflow', [++$generation]);
    }
}
