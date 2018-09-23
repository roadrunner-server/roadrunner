<?php

use Psr\Http\Message\ResponseInterface;
use Psr\Http\Message\ServerRequestInterface;

function handleRequest(ServerRequestInterface $req, ResponseInterface $resp): ResponseInterface
{
    throw new Error("error");
}