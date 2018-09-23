<?php

use \Psr\Http\Message\ServerRequestInterface;
use \Psr\Http\Message\ResponseInterface;

function handleRequest(ServerRequestInterface $req, ResponseInterface $resp): ResponseInterface
{
    error_log(strtoupper($req->getQueryParams()['hello']));

    $resp->getBody()->write(strtoupper($req->getQueryParams()['hello']));
    return $resp->withStatus(201);
}