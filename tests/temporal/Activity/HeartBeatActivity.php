<?php

namespace Temporal\Tests\Activity;

use Temporal\Activity;
use Temporal\Activity\ActivityInterface;
use Temporal\Activity\ActivityMethod;
use Temporal\Roadrunner\Internal\Error;

#[ActivityInterface(prefix: "HeartBeatActivity.")]
class HeartBeatActivity
{
    #[ActivityMethod]
    public function doSomething(
        int $value
    ): string {
        Activity::heartbeat(['value' => $value]);
        sleep($value);
        return 'OK';
    }

    #[ActivityMethod]
    public function slow(
        string $value
    ): string {
        for ($i = 0; $i < 5; $i++) {
            Activity::heartbeat(['value' => $i]);
            sleep(1);
        }

        return 'OK';
    }

    #[ActivityMethod]
    public function something(
        string $value
    ): string {
        Activity::heartbeat(['value' => $value]);
        sleep($value);
        return 'OK';
    }

    #[ActivityMethod]
    public function failedActivity(
        int $value
    ): string {
        Activity::heartbeat(['value' => $value]);
        if (Activity::getInfo()->attempt === 1) {
            throw new \Error("failed");
        }

        if (!is_array(Activity::getHeartbeatDetails())) {
            throw new \Error("no heartbeat details");
        }

        return 'OK!';
    }
}