package main

import (
	"../pb"
	"context"
	"flag"
	"github.com/golang/protobuf/ptypes"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"log"
	"time"
)

const (
	currentToken = "aaa"
	wrongToken   = "aaa1"
	authority    = "aaaa"
)

func main() {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	// WithAuthority会将传入值设置到请求头的:authority字段中,服务端可以检查此字段(如果未用安全套接字,此字段将以明文进行传输)
	opts = append(opts,
		grpc.WithAuthority(authority),
		grpc.WithStreamInterceptor(StreamClientInterceptor),
		grpc.WithUnaryInterceptor(UnaryClientInterceptor),
	)
	//opts = append(opts, grpc.WithPerRPCCredentials(&auth))
	target := flag.String("t", "127.0.0.1:1234", "server address(ip:port)")
	conn, err := grpc.Dial(*target, opts...)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	client := pb.NewTryServiceClient(conn)
	// cancelCtx传入到Notify中,可在需要断开流时取消此cancelCtx
	cancelCtx, cancelFunc := context.WithCancel(context.Background())
	// 这是通过外传metadata存放token的一种示例,stream是用于流
	authStreamMd := metadata.Pairs("token", currentToken)
	ctx := metadata.NewOutgoingContext(cancelCtx, authStreamMd)
	// 双向流模式
	nc, err := client.Notify(ctx, &pb.TryRequest{Id: "xunj"})
	if err != nil {
		log.Fatal(err)
	}
	timeout := time.NewTimer(time.Hour).C
	go func() {
		// 超时主动断开流
		<-timeout
		cancelFunc()
	}()
	go func() {
		for {
			msg, err := nc.Recv()
			if err != nil {
				log.Println(err)
				return
			}
			// 解析protobuf的Any类型
			addr := pb.Address{}
			err = ptypes.UnmarshalAny(msg.Details, &addr)
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("Recv:%s\tDetail:%s:%d", msg.Content, addr.GetIP(), addr.GetPort())
		}
	}()
	// 这是通过外传metadata存放token的一种示例,unary是用于一元调用
	authUnaryMd := metadata.Pairs("token", wrongToken)
	ctx1 := metadata.NewOutgoingContext(context.Background(), authUnaryMd)
	// 一元调用
	last, err := client.LastNotify(ctx1, &pb.TryRequest{Id: "xunj"})
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Last:" + last.Content)
	<-make(chan bool)
}

