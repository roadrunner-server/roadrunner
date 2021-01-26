<?php

namespace Temporal\Tests\Workflow;

use Temporal\Workflow;
use Temporal\Workflow\WorkflowMethod;

#[Workflow\WorkflowInterface]
class PeriodicWorkflow
{
    #[WorkflowMethod(name: 'PeriodicWorkflow')]
    public function handler()
    {
        error_log("GOT SOMETHING" . print_r(Workflow::getLastCompletionResult(), true));

        // todo: get last completion result
        return 'OK';
    }
}
