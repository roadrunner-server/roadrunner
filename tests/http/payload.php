<?php

use Psr\Http\Message\ResponseInterface;
use Psr\Http\Message\ServerRequestInterface;

function handleRequest(ServerRequestInterface $req, ResponseInterface $resp): ResponseInterface
{
    if ( $req->getHeaderLine("Content-Type") != 'application/json' ) {
        $resp->getBody()->write("invalid content-type");
        return $resp;
    }

    // we expect json body
    $p = json_decode($req->getBody(), true);
    $resp->getBody()->write(json_encode(array_flip($p)));

    return $resp;
}