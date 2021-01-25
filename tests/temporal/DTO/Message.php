<?php


namespace Temporal\Tests\DTO;

class Message
{
    public string $message;

    public function __construct(string $message)
    {
        $this->message = $message;
    }
}