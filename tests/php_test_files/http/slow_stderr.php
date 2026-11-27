<?php

declare(strict_types=1);

use Psr\Http\Message\ResponseInterface;
use Psr\Http\Message\ServerRequestInterface;

function handleRequest(ServerRequestInterface $req, ResponseInterface $resp): ResponseInterface
{
    $path = $req->getUri()->getPath();
    if ($path === '/slow') {
        $ready = $req->getQueryParams()['ready'] ?? '';
        if (is_string($ready) && $ready !== '') {
            file_put_contents($ready, '1');
        }
        sleep(1);
        fwrite(STDERR, json_encode([
            'message' => 'request finished',
            'level_name' => 'INFO',
            'path' => $path,
        ], JSON_UNESCAPED_SLASHES) . "\n");
    }

    $resp->getBody()->write('done');
    return $resp->withStatus(200)->withHeader('Content-Length', '4');
}
