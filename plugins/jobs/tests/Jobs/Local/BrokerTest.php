<?php

/**
 * Spiral Framework.
 *
 * @license   MIT
 * @author    Anton Titov (Wolfy-J)
 */

declare(strict_types=1);

namespace Spiral\Jobs\Tests\Local;

use Spiral\Jobs\Tests\BaseTest;

class BrokerTest extends BaseTest
{
    public const JOB       = Job::class;
    public const ERROR_JOB = ErrorJob::class;
}
