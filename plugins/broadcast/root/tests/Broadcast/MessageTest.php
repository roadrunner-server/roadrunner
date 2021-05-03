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
use Spiral\Broadcast\Message;

class MessageTest extends TestCase
{
    public function testSerialize(): void
    {
        $m = new Message('topic', ['hello' => 'world']);
        $this->assertSame('{"topic":"topic","payload":{"hello":"world"}}', json_encode($m));
    }
}
