<?php


namespace Temporal\Tests\Workflow;

use Temporal\Activity\ActivityOptions;
use Temporal\Tests\Activity\SimpleActivity;
use Temporal\Workflow;
use Temporal\Workflow\SignalMethod;
use Temporal\Workflow\WorkflowMethod;

#[Workflow\WorkflowInterface]
class LoopWorkflow
{
    private array $values = [];
    private array $result = [];
    private $simple;

    public function __construct()
    {
        $this->simple = Workflow::newActivityStub(
            SimpleActivity::class,
            ActivityOptions::new()->withStartToCloseTimeout(5)
        );
    }

    #[SignalMethod]
    public function addValue(
        string $value
    ) {
        $this->values[] = $value;
    }

    #[WorkflowMethod(name: 'LoopWorkflow')]
    public function run(
        int $count
    ) {
        while (true) {
            yield Workflow::await(fn() => $this->values !== []);
            $value = array_shift($this->values);

            $this->result[] = yield $this->simple->echo($value);

            if (count($this->result) === $count) {
                break;
            }
        }

        return $this->result;
    }
}
