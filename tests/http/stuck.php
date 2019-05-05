<?php

use \Psr\Http\Message\ServerRequestInterface;
use \Psr\Http\Message\ResponseInterface;

function handleRequest(ServerRequestInterface $req, ResponseInterface $resp): ResponseInterface
{
    sleep(10);
    $resp->getBody()->write(strtoupper($req->getQueryParams()['hello']));
    return $resp->withStatus(201);
}