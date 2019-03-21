package main

import (
	"context"
	"google.golang.org/grpc"
	"log"
)

func UnaryClientInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	log.Printf("Send rpc \"%+v\":(%+v)", method, req)
	err := invoker(ctx, method, req, reply, cc, opts...)
	log.Printf("Recive reply: \"%+v\":%+v", method, reply)
	return err
}

func StreamClientInterceptor(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	log.Printf("treamClientInterceptor \"%+v\"%+v", method, desc)
	clientStream, err := streamer(ctx, desc, cc, method, opts...)
	return clientStream, err
}
