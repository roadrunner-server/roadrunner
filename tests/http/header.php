<?php

use Psr\Http\Message\ResponseInterface;
use Psr\Http\Message\ServerRequestInterface;

function handleRequest(ServerRequestInterface $req, ResponseInterface $resp): ResponseInterface
{
    $resp->getBody()->write(strtoupper($req->getHeaderLine('input')));

    return $resp->withAddedHeader("Header", $req->getQueryParams()['hello']);
}