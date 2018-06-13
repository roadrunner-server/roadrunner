<?php

use Psr\Http\Message\ResponseInterface;
use Psr\Http\Message\ServerRequestInterface;

function handleRequest(ServerRequestInterface $req, ResponseInterface $resp): ResponseInterface
{
    $files = $req->getUploadedFiles();
    array_walk_recursive($files, function (&$v) {
        /**
         * @var \Psr\Http\Message\UploadedFileInterface $v
         */

        if ($v->getError()) {
            $v = [
                'name'  => $v->getClientFilename(),
                'size'  => $v->getSize(),
                'mime'  => $v->getClientMediaType(),
                'error' => $v->getError(),
            ];
        } else {
            $v = [
                'name'  => $v->getClientFilename(),
                'size'  => $v->getSize(),
                'mime'  => $v->getClientMediaType(),
                'error' => $v->getError(),
                'md5'   => md5($v->getStream()->__toString()),
            ];
        }
    });

    $resp->getBody()->write(json_encode($files, JSON_UNESCAPED_SLASHES));

    return $resp;
}