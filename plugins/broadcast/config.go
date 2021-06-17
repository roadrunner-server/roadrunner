package broadcast

/*
websockets: # <----- one of possible subscribers
  path: /ws
  broker: default # <------ broadcast broker to use --------------- |
                                                                    |  match
broadcast: # <-------- broadcast entry point plugin                 |
  default: # <----------------------------------------------------- |
     driver: redis
  test:
     driver: memory

*/

// Config ...
type Config struct {
	Data map[string]interface{} `mapstructure:"broadcast"`
}
