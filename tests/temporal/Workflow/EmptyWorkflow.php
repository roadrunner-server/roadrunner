<?php

namespace Temporal\Tests\Workflow;

use Temporal\Workflow\WorkflowMethod;
use Temporal\Workflow;

#[Workflow\WorkflowInterface]
class EmptyWorkflow
{
    #[WorkflowMethod]
    public function handler()
    {
        return 42;
    }
}
