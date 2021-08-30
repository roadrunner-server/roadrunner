### Job lifecycle

There are several boltdb buckets:

1. `PushBucket` - used for pushed jobs via RPC.
2. `InQueueBucket` - when the job consumed from the `PushBucket`, in the same transaction, it copied into the priority queue and
get into the `InQueueBucket` waiting to acknowledgement.
3. `DelayBucket` - used for delayed jobs. RFC3339 used as a timestamp to track delay expiration.

``
