package client

import (
	"context"
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
)

var (
	ErrNotFoundClient = errors.New("not found grpc client service")
	ErrConnShutdown   = errors.New("grpc connection has closed")

	defaultPoolSize                    = 10
	defaultDialTimeout                 = 10 * time.Second
	defaultKeepAlive                   = 30 * time.Second
	defaultKeepAliveTimeout            = 10 * time.Second
	defaultBackoffMaxDelay             = 3 * time.Second
	defaultMaxSendMsgSize              = 4 << 20
	defaultMaxMaxRecvMsgSize           = 4 << 20
	defaultInitialWindowSize     int32 = 4 << 20
	defaultInitialConnWindowSize int32 = 4 << 20
)

type ClientOption struct {
	PoolSize         int
	DialTimeOut      time.Duration
	KeepAlive        time.Duration
	KeepAliveTimeout time.Duration
}

type ClientPool struct {
	endpoint string
	next     int64
	cap      int64

	option *ClientOption
	conns  []*grpc.ClientConn
	sync.Mutex
}

func (cc *ClientPool) getConn() (*grpc.ClientConn, error) {
	var (
		idx  int64
		next int64
		err  error
	)

	next = atomic.AddInt64(&cc.next, 1)
	idx = next % cc.cap
	conn := cc.conns[idx]
	if conn != nil && cc.checkState(conn) == nil {
		return conn, nil
	}

	if conn != nil {
		conn.Close()
	}

	cc.Lock()
	defer cc.Unlock()

	// 双检, 防止已经初始化
	conn = cc.conns[idx]
	if conn != nil && cc.checkState(conn) == nil {
		return conn, nil
	}

	conn, err = cc.connect()
	if err != nil {
		return nil, err
	}

	cc.conns[idx] = conn
	return conn, nil
}

func (cc *ClientPool) checkState(conn *grpc.ClientConn) error {
	state := conn.GetState()
	switch state {
	case connectivity.TransientFailure, connectivity.Shutdown:
		return ErrConnShutdown
	}

	return nil
}

func (cc *ClientPool) connect() (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(context.TODO(), cc.option.DialTimeOut)
	defer cancel()
	conn, err := grpc.DialContext(ctx,
		cc.endpoint,
		//grpc.WithBlock(),
		//grpc.WithConnectParams(grpc.ConnectParams{
		//	Backoff:           backoff.Config{MaxDelay: 8*time.Second},
		//	MinConnectTimeout: 0,
		//}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(UnaryClientInterceptor),
		grpc.WithInitialWindowSize(defaultInitialWindowSize),
		grpc.WithInitialConnWindowSize(defaultInitialConnWindowSize),
		grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(defaultMaxSendMsgSize)),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(defaultMaxMaxRecvMsgSize)),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                cc.option.KeepAlive,
			Timeout:             cc.option.KeepAliveTimeout,
			PermitWithoutStream: true,
		}))
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func (cc *ClientPool) Close() {
	cc.Lock()
	defer cc.Unlock()

	for _, conn := range cc.conns {
		if conn == nil {
			continue
		}
		conn.Close()
	}
}

func newClientPoolWithOption(endpoint string, option *ClientOption) *ClientPool {
	if (option.PoolSize) <= 0 {
		option.PoolSize = defaultPoolSize
	}

	if option.DialTimeOut <= 0 {
		option.DialTimeOut = defaultDialTimeout
	}

	if option.KeepAlive <= 0 {
		option.KeepAlive = defaultKeepAlive
	}

	if option.KeepAliveTimeout <= 0 {
		option.KeepAliveTimeout = defaultKeepAliveTimeout
	}

	return &ClientPool{
		endpoint: endpoint,
		option:   option,
		cap:      int64(option.PoolSize),
		conns:    make([]*grpc.ClientConn, option.PoolSize),
	}
}

type ServiceClientPool struct {
	option   *ClientOption
	services map[string][]string
	clients  map[string]*ClientPool
}

func NewServiceClientPool(option *ClientOption) *ServiceClientPool {
	return &ServiceClientPool{
		option:   option,
		services: make(map[string][]string),
	}
}

func (sc *ServiceClientPool) Start() {
	var clients = make(map[string]*ClientPool, len(sc.services))
	for endpoint, srvNameArr := range sc.services {
		cc := newClientPoolWithOption(endpoint, sc.option)
		for _, srv := range srvNameArr {
			clients[srv] = cc
		}
	}

	sc.clients = clients
	scp = sc
}

func (sc *ServiceClientPool) SetServices(endpoint string, services ...string) {
	if len(services) == 0 {
		return
	}
	sc.services[endpoint] = append(sc.services[endpoint], services...)
}

func (sc *ServiceClientPool) GetClientWithFullMethod(fullMethod string) (*grpc.ClientConn, error) {
	sn := sc.spiltFullMethod(fullMethod)
	return sc.GetClient(sn)
}

func (sc *ServiceClientPool) GetClient(serviceName string) (*grpc.ClientConn, error) {
	cc, ok := sc.clients[serviceName]
	if !ok {
		return nil, ErrNotFoundClient
	}

	return cc.getConn()
}

func (sc *ServiceClientPool) Close(serviceName string) {
	cc, ok := sc.clients[serviceName]
	if !ok {
		return
	}

	cc.Close()
}

func (sc *ServiceClientPool) CloseAll() {
	for _, client := range sc.clients {
		client.Close()
	}
}

func (sc *ServiceClientPool) spiltFullMethod(fullMethod string) string {
	var arr []string
	arr = strings.Split(fullMethod, "/")
	if len(arr) != 3 {
		return ""
	}

	return arr[1]
}

func (sc *ServiceClientPool) Invoke(
	ctx context.Context,
	fullMethod string,
	headers map[string]string,
	args interface{},
	reply interface{},
	opts ...grpc.CallOption,
) error {
	var md metadata.MD
	serviceName := sc.spiltFullMethod(fullMethod)
	conn, err := sc.GetClient(serviceName)
	if err != nil {
		return err
	}

	md, exist := metadata.FromOutgoingContext(ctx)
	if exist {
		md = md.Copy()
	} else {
		md = metadata.MD{}
	}

	for k, v := range headers {
		md.Set(k, v)
	}

	ctx = metadata.NewOutgoingContext(ctx, md)
	return conn.Invoke(ctx, fullMethod, args, reply, opts...)
}

var scp *ServiceClientPool

func NewDefaultPool() *ServiceClientPool {
	co := ClientOption{
		PoolSize:         defaultPoolSize,
		DialTimeOut:      defaultDialTimeout,
		KeepAlive:        defaultKeepAlive,
		KeepAliveTimeout: defaultKeepAliveTimeout,
	}
	return NewServiceClientPool(&co)
}

func GetPool() *ServiceClientPool {
	return scp
}
