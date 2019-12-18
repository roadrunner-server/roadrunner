<?php

use \Psr\Http\Message\ServerRequestInterface;
use \Psr\Http\Message\ResponseInterface;
use Spiral\Goridge\RPC;
use Spiral\Goridge\SocketRelay;
use Spiral\RoadRunner\RPCLogger;

function handleRequest(ServerRequestInterface $req, ResponseInterface $resp): ResponseInterface
{
    $rpc = new RPC(new SocketRelay("127.0.0.1", 6001));
    $logger = new RPCLogger($rpc);

    $level = $req->getQueryParams()['level'] ?? 'warning';
    $message = $req->getQueryParams()['message'] ?? 'default message';
    $context = (array) ($req->getQueryParams()['fields'] ?? []);

    $logger->log($level, $message, $context);

    return $resp->withStatus(201);
}
