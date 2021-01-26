<?php

namespace Temporal\Tests\Workflow;

use Temporal\Activity\ActivityOptions;
use Temporal\DataConverter\Bytes;
use Temporal\Workflow;
use Temporal\Workflow\WorkflowMethod;

#[Workflow\WorkflowInterface]
class BinaryWorkflow
{
    #[WorkflowMethod(name: 'BinaryWorkflow')]
    public function handler(
        Bytes $input
    ): iterable {
        $opts = ActivityOptions::new()->withStartToCloseTimeout(5);

        return yield Workflow::executeActivity('SimpleActivity.sha512', [$input], $opts);
    }
}
