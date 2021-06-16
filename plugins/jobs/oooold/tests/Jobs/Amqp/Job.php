<?php

/**
 * Spiral Framework.
 *
 * @license   MIT
 * @author    Anton Titov (Wolfy-J)
 */

declare(strict_types=1);

namespace Spiral\Jobs\Tests\Amqp;

use Spiral\Jobs\JobHandler;

class Job extends JobHandler
{
    public const JOB_FILE = __DIR__ . '/../../local.job';

    public function invoke(string $id, array $payload): void
    {
        file_put_contents(self::JOB_FILE, json_encode(
            $payload + compact('id')
        ));
    }
}
