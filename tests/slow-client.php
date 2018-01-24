<?php

use Spiral\Goridge;

ini_set('display_errors', 'stderr');
require dirname(__DIR__) . "/vendor/autoload.php";

if (count($argv) < 3) {
    die("need 2 arguments");
}

list($test, $goridge, $bootDelay, $shutdownDelay) = [$argv[1], $argv[2], $argv[3], $argv[4]];

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

usleep($bootDelay * 1000);
require_once sprintf("%s/%s.php", __DIR__, $test);
usleep($shutdownDelay * 1000);