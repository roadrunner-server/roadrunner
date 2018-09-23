<?php

use Psr\Http\Message\ResponseInterface;
use Psr\Http\Message\ServerRequestInterface;

function handleRequest(ServerRequestInterface $req, ResponseInterface $resp): ResponseInterface
{

    $data = $req->getParsedBody();

    ksort($data);
    ksort($data['arr']);
    ksort($data['arr']['x']['y']);

    $resp->getBody()->write(json_encode($data));

    return $resp;
}