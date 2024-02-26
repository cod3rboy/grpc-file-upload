package interceptors

import (
	"log"
	"time"

	"google.golang.org/grpc"
)

type streamLogger struct {
	grpc.ServerStream
}

func (s *streamLogger) RecvMsg(m interface{}) error {
	log.Printf("====== [Server Stream Interceptor Logger] ======\nReceive a message (Type: %T) at %s", m, time.Now().Format(time.RFC3339))
	return s.ServerStream.RecvMsg(m)
}

func (s *streamLogger) SendMsg(m interface{}) error {
	log.Printf("====== [Server Stream Interceptor Logger] ======\nSend a message (Type: %T) at %s", m, time.Now().Format(time.RFC3339))
	return s.ServerStream.SendMsg(m)
}

func newStreamLogger(s grpc.ServerStream) grpc.ServerStream {
	return &streamLogger{s}
}

func StreamLogInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	log.Println("====== [Server Stream Interceptor] ", info.FullMethod)
	err := handler(srv, newStreamLogger(ss))
	if err != nil {
		log.Printf("RPC failed with error: %v", err)
	}
	return err
}
