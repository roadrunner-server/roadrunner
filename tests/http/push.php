<?php

use \Psr\Http\Message\ServerRequestInterface;
use \Psr\Http\Message\ResponseInterface;

function handleRequest(ServerRequestInterface $req, ResponseInterface $resp): ResponseInterface
{
    $resp->getBody()->write(strtoupper($req->getQueryParams()['hello']));
    return $resp->withAddedHeader("Http2-Push", __FILE__)->withStatus(201);
}
