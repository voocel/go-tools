package server

import (
	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
)

type GrpcServerBuilder struct {
	options           []grpc.ServerOption
	enabledReflection bool
	enabledPrometheus bool
	enabledHealth     bool
}

func (b *GrpcServerBuilder) EnableReflection() {
	b.enabledReflection = true
}

func (b *GrpcServerBuilder) EnablePrometheus() {
	b.enabledPrometheus = true
}

func (b *GrpcServerBuilder) EnabledHealth()  {
	b.enabledHealth = true
}

func (b *GrpcServerBuilder) AddOption(opt grpc.ServerOption) {
	b.options = append(b.options, opt)
}

func (b *GrpcServerBuilder) SetServerParams(params keepalive.ServerParameters) {
	keepAlive := grpc.KeepaliveParams(params)
	b.AddOption(keepAlive)
}

func (b *GrpcServerBuilder) SetStreamInterceptors(interceptors ...grpc.StreamServerInterceptor) {
	chain := grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(interceptors...))
	b.AddOption(chain)
}

func (b *GrpcServerBuilder) SetUnaryInterceptors(interceptors ...grpc.UnaryServerInterceptor) {
	chain := grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(interceptors...))
	b.AddOption(chain)
}

func (b *GrpcServerBuilder) SetTLSCert(serverKeyPath, serverPemPath, caPemPath string) {
	cred := setCert(serverKeyPath, serverPemPath, caPemPath)
	b.AddOption(cred)
}

func (b *GrpcServerBuilder) Build() GrpcServer {
	srv := grpc.NewServer(b.options...)
	if b.enabledReflection {
		reflection.Register(srv)
	}
	if b.enabledPrometheus {
		grpc_prometheus.EnableHandlingTimeHistogram()
		grpc_prometheus.Register(srv)
	}
	if b.enabledHealth {
		h := health.NewServer()
		grpc_health_v1.RegisterHealthServer(srv, h)
		serviceName := "grpc_service_name"
		h.SetServingStatus(serviceName, grpc_health_v1.HealthCheckResponse_SERVING)
	}
	return &grpcServer{srv, nil}
}
