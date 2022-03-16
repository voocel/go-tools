package server

import (
	"google.golang.org/grpc"
	"net"
	"os"
	"os/signal"
	"syscall"
)

type GrpcServer interface {
	Start(addr string) error
	StartWithListener(l net.Listener)
	Await(func())
	RegisterService(func(*grpc.Server))
	Serve()
	Stop()
}

type grpcServer struct {
	server   *grpc.Server
	listener net.Listener
}

func (s *grpcServer) Start(addr string) (err error) {
	s.listener, err = net.Listen("tcp", addr)
	if err != nil {
		return
	}
	go s.Serve()
	return
}

func (s *grpcServer) StartWithListener(l net.Listener) {
	s.listener = l
	go s.Serve()
	return
}

func (s *grpcServer) Await(hook func()) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	<-c
	s.Stop()
	if hook != nil {
		hook()
	}
}

func (s *grpcServer) RegisterService(reg func(*grpc.Server)) {
	reg(s.server)
}

func (s *grpcServer) Serve() {
	if err := s.server.Serve(s.listener); err != nil {
		panic(err)
	}
}

func (s *grpcServer) Stop() {
	s.server.GracefulStop()
	s.listener.Close()
}
