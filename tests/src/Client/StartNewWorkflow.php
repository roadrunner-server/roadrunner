<?php


namespace Temporal\Tests\Client;

use Temporal\Client;
use Temporal\Tests\Workflow\SimpleDTOWorkflow;

use function Symfony\Component\String\s;

class StartNewWorkflow
{
    private $stub;

    public function __construct(Client\ClientInterface $client)
    {
        $this->stub = $client->newWorkflowStub(SimpleDTOWorkflow::class);
    }

    public function __invoke()
    {
    }
}
