package proxy

// import (
// 	"testing"
// 	"time"

// 	"github.com/sirupsen/logrus"
// 	"github.com/sirupsen/logrus/hooks/test"
// 	"github.com/stretchr/testify/assert"
// 	"golang.org/x/net/context"
// 	"google.golang.org/grpc"
// 	"google.golang.org/grpc/codes"
// 	"google.golang.org/grpc/metadata"
// 	"google.golang.org/grpc/status"
// )

// const addr = "localhost:9080"

// func Test_Proxy_Error(t *testing.T) {
// 	logger, _ := test.NewNullLogger()
// 	logger.SetLevel(logrus.DebugLevel)

// 	c := service.NewContainer(logger)
// 	c.Register(ID, &Service{})

// 	assert.NoError(t, c.Init(&testCfg{
// 		grpcCfg: `{
// 			"listen": "tcp://:9080",
// 			"tls": {
// 				"key": "tests/server.key",
// 				"cert": "tests/server.crt"
// 			},
// 			"proto": "tests/test.proto",
// 			"workers":{
// 				"command": "php tests/worker.php",
// 				"relay": "pipes",
// 				"pool": {
// 					"numWorkers": 1,
// 					"allocateTimeout": 10,
// 					"destroyTimeout": 10
// 				}
// 			}
// 	}`,
// 	}))

// 	s, st := c.Get(ID)
// 	assert.NotNil(t, s)
// 	assert.Equal(t, service.StatusOK, st)

// 	// should do nothing
// 	s.(*Service).Stop()

// 	go func() { assert.NoError(t, c.Serve()) }()
// 	time.Sleep(time.Millisecond * 100)
// 	defer c.Stop()

// 	cl, cn := getClient(addr)
// 	defer cn.Close()

// 	_, err := cl.Throw(context.Background(), &tests.Message{Msg: "notFound"})

// 	assert.Error(t, err)
// 	se, _ := status.FromError(err)
// 	assert.Equal(t, "nothing here", se.Message())
// 	assert.Equal(t, codes.NotFound, se.Code())

// 	_, errWithDetails := cl.Throw(context.Background(), &tests.Message{Msg: "withDetails"})

// 	assert.Error(t, errWithDetails)
// 	statusWithDetails, _ := status.FromError(errWithDetails)
// 	assert.Equal(t, "main exception message", statusWithDetails.Message())
// 	assert.Equal(t, codes.InvalidArgument, statusWithDetails.Code())

// 	details := statusWithDetails.Details()

// 	detailsMessageForException := details[0].(*tests.DetailsMessageForException)

// 	assert.Equal(t, detailsMessageForException.Code, uint64(1))
// 	assert.Equal(t, detailsMessageForException.Message, "details message")
// }

// func Test_Proxy_Metadata(t *testing.T) {
// 	logger, _ := test.NewNullLogger()
// 	logger.SetLevel(logrus.DebugLevel)

// 	c := service.NewContainer(logger)
// 	c.Register(ID, &Service{})

// 	assert.NoError(t, c.Init(&testCfg{
// 		grpcCfg: `{
// 			"listen": "tcp://:9080",
// 			"tls": {
// 				"key": "tests/server.key",
// 				"cert": "tests/server.crt"
// 			},
// 			"proto": "tests/test.proto",
// 			"workers":{
// 				"command": "php tests/worker.php",
// 				"relay": "pipes",
// 				"pool": {
// 					"numWorkers": 1,
// 					"allocateTimeout": 10,
// 					"destroyTimeout": 10
// 				}
// 			}
// 	}`,
// 	}))

// 	s, st := c.Get(ID)
// 	assert.NotNil(t, s)
// 	assert.Equal(t, service.StatusOK, st)

// 	// should do nothing
// 	s.(*Service).Stop()

// 	go func() { assert.NoError(t, c.Serve()) }()
// 	time.Sleep(time.Millisecond * 100)
// 	defer c.Stop()

// 	cl, cn := getClient(addr)
// 	defer cn.Close()

// 	ctx := metadata.AppendToOutgoingContext(context.Background(), "key", "proxy-value")
// 	var header metadata.MD
// 	out, err := cl.Info(
// 		ctx,
// 		&tests.Message{Msg: "MD"},
// 		grpc.Header(&header),
// 		grpc.WaitForReady(true),
// 	)
// 	assert.Equal(t, []string{"bar"}, header.Get("foo"))
// 	assert.NoError(t, err)
// 	assert.Equal(t, `["proxy-value"]`, out.Msg)
// }
