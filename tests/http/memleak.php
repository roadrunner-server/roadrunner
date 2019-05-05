<?php

use Psr\Http\Message\ResponseInterface;
use Psr\Http\Message\ServerRequestInterface;

$mem = '';

function handleRequest(ServerRequestInterface $req, ResponseInterface $resp): ResponseInterface
{
   $mem .= str_repeat(" ", 1024*1024);
   return $resp;
}