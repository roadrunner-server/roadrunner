<?php

declare(strict_types=1);

namespace Temporal\Tests\Workflow;

use Temporal\Activity\ActivityOptions;
use Temporal\Promise;
use Temporal\Workflow;
use Temporal\Workflow\WorkflowMethod;
use Temporal\Tests\Activity\SimpleActivity;

#[Workflow\WorkflowInterface]
class ParallelScopesWorkflow
{
    #[WorkflowMethod(name: 'ParallelScopesWorkflow')]
    public function handler(string $input)
    {
        $simple = Workflow::newActivityStub(
            SimpleActivity::class,
            ActivityOptions::new()->withStartToCloseTimeout(5)
        );

        $a = Workflow::newCancellationScope(function () use ($simple, $input) {
            return yield $simple->echo($input);
        });

        $b = Workflow::newCancellationScope(function () use ($simple, $input) {
            return yield $simple->lower($input);
        });

        [$ra, $rb] = yield Promise::all([$a, $b]);

        return sprintf('%s|%s|%s', $ra, $input, $rb);
    }
}
