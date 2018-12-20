<?php

use Psr\Http\Message\ResponseInterface;
use Psr\Http\Message\ServerRequestInterface;

function handleRequest(ServerRequestInterface $req, ResponseInterface $resp): ResponseInterface
{
    $resp->getBody()->write($_SERVER['HTTP_USER_AGENT']);
    return $resp;
}