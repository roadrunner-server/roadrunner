<?php

namespace Temporal\Tests\Activity;

use Temporal\Activity\ActivityInterface;
use Temporal\Activity\ActivityMethod;
use Temporal\Api\Common\V1\WorkflowExecution;
use Temporal\DataConverter\Bytes;
use Temporal\Tests\DTO\Message;
use Temporal\Tests\DTO\User;

#[ActivityInterface(prefix: "SimpleActivity.")]
class SimpleActivity
{
    #[ActivityMethod]
    public function echo(
        string $input
    ): string {
        return strtoupper($input);
    }

    #[ActivityMethod]
    public function lower(
        string $input
    ): string {
        return strtolower($input);
    }

    #[ActivityMethod]
    public function greet(
        User $user
    ): Message {
        return new Message(sprintf("Hello %s <%s>", $user->name, $user->email));
    }

    #[ActivityMethod]
    public function slow(
        string $input
    ): string {
        sleep(2);

        return strtolower($input);
    }

    #[ActivityMethod]
    public function sha512(
        Bytes $input
    ): string {
        return hash("sha512", ($input->getData()));
    }

    public function updateRunID(WorkflowExecution $e): WorkflowExecution
    {
        $e->setRunId('updated');
        return $e;
    }

    #[ActivityMethod]
    public function fail()
    {
        throw new \Error("failed activity");
    }
}