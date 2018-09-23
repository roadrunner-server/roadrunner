<?php

use Spiral\Goridge;

ini_set('display_errors', 'stderr');
require dirname(__DIR__) . "/vendor/autoload.php";

if (count($argv) < 3) {
    die("need 2 arguments");
}

list($test, $goridge) = [$argv[1], $argv[2]];

switch ($goridge) {
    case "pipes":
        $relay = new Goridge\StreamRelay(STDIN, STDOUT);
        break;

    case "tcp":
        $relay = new Goridge\SocketRelay("localhost", 9007);
        break;

    case "unix":
        $relay = new Goridge\SocketRelay(
            "sock.unix",
            null,
            Goridge\SocketRelay::SOCK_UNIX
        );
        break;

    default:
        die("invalid protocol selection");
}

require_once sprintf("%s/%s.php", __DIR__, $test);

sleep(10);