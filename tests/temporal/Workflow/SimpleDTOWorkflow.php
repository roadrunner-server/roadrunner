<?php

declare(strict_types=1);

namespace Temporal\Tests\Workflow;

use Temporal\Activity\ActivityOptions;
use Temporal\Workflow;
use Temporal\Workflow\WorkflowMethod;
use Temporal\Tests\Activity\SimpleActivity;
use Temporal\Tests\DTO\Message;
use Temporal\Tests\DTO\User;

#[Workflow\WorkflowInterface]
class SimpleDTOWorkflow
{
    #[WorkflowMethod(name: 'SimpleDTOWorkflow')]//, returnType: Message::class)]
    public function handler(
        User $user
    ) {
        $simple = Workflow::newActivityStub(
            SimpleActivity::class,
            ActivityOptions::new()
                ->withStartToCloseTimeout(5)
        );

        $value = yield $simple->greet($user);

        if (!$value instanceof Message) {
            return "FAIL";
        }

        return $value;
    }
}
