<?php
/**
 * Sample GRPC PHP server.
 */

use Spiral\GRPC\ContextInterface;
use Service\EchoInterface;
use Service\Message;

class EchoService implements EchoInterface
{
    public function Ping(ContextInterface $ctx, Message $in): Message
    {
        $out = new Message();
        return $out->setMsg(strtoupper($in->getMsg()));
    }
}