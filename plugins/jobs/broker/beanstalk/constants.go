package beanstalk

import "time"

// WatchThrottleLimit is used to limit reconnection occurrence in watch function
const WatchThrottleLimit = time.Second