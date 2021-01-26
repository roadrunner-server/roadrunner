<?php

/**
 * This file is part of Temporal package.
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

namespace Temporal\Tests\Workflow;

use Temporal\Workflow;
use Temporal\Workflow\WorkflowMethod;

#[Workflow\WorkflowInterface]
class QueryWorkflow
{
    private int $counter = 0;

    #[Workflow\SignalMethod(name: "add")]
    public function add(
        int $value
    ) {
        $this->counter += $value;
    }

    #[Workflow\QueryMethod(name: "get")]
    public function get(): int
    {
        return $this->counter;
    }

    #[WorkflowMethod]
    public function handler()
    {
        // collect signals during one second
        yield Workflow::timer(1);

        return $this->counter;
    }
}
