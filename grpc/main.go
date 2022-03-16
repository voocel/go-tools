package main

import (
	"google.golang.org/grpc"
	"grpc/client"
	"grpc/server"
	"log"
)

func main()  {
  // run grpc server
	RunServer()
}

func RunClient()  {
	pool := client.NewDefaultPool()
	pool.SetServices("127.0.0.1:6868", "hello.HelloServer")
	pool.Start()
}

func RunServer()  {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	builder := server.GrpcServerBuilder{}
	builder.EnableReflection()

	s := builder.Build()
	s.RegisterService(serviceRegister)
	err := s.Start(":6868")
	if err != nil {
		log.Fatal(err)
	}

	s.Await(func() {
		log.Print("Shutting down the server")
	})
}

func serviceRegister(s *grpc.Server) {
	// helloPB.RegisterHelloServer(s, &helloService{})
	// helloPB2.RegisterHelloServer2(s, &helloService2{})
}
