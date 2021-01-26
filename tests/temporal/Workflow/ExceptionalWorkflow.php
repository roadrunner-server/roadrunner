<?php

namespace Temporal\Tests\Workflow;

use Temporal\Activity\ActivityOptions;
use Temporal\Tests\Activity\SimpleActivity;
use Temporal\Workflow;
use Temporal\Workflow\WorkflowMethod;

#[Workflow\WorkflowInterface]
class ExceptionalWorkflow
{
    #[WorkflowMethod(name: 'ExceptionalWorkflow')]
    public function handler()
    {
        throw new \RuntimeException("workflow error");
    }
}
