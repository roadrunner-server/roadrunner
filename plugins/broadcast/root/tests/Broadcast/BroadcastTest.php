<?php

/**
 * Spiral Framework.
 *
 * @license   MIT
 * @author    Anton Titov (Wolfy-J)
 */

declare(strict_types=1);

namespace Spiral\Broadcast\Tests;

use PHPUnit\Framework\TestCase;
use Spiral\Broadcast\Broadcast;
use Spiral\Broadcast\Exception\BroadcastException;
use Spiral\Broadcast\Message;
use Spiral\Goridge\RPC;
use Spiral\Goridge\SocketRelay;

class BroadcastTest extends TestCase
{
    public function testBroadcast(): void
    {
        $rpc = new RPC(new SocketRelay('localhost', 6001));
        $br = new Broadcast($rpc);

        $br->publish(
            new Message('tests/topic', 'hello'),
            new Message('tests/123', ['key' => 'value'])
        );

        while (filesize(__DIR__ . '/../log.txt') < 40) {
            clearstatcache(true, __DIR__ . '/../log.txt');
            usleep(1000);
        }

        clearstatcache(true, __DIR__ . '/../log.txt');
        $content = file_get_contents(__DIR__ . '/../log.txt');

        $this->assertSame('tests/topic: "hello"
tests/123: {"key":"value"}
', $content);
    }

    public function testBroadcastException(): void
    {
        $rpc = new RPC(new SocketRelay('localhost', 6002));
        $br = new Broadcast($rpc);

        $this->expectException(BroadcastException::class);
        $br->publish(
            new Message('topic', 'hello')
        );
    }
}
