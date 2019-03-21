//go:generate protoc --go_out=plugins=grpc:../pb -I ../pb ../pb/try.proto

package main

import (
	"../pb"
	"context"
	"flag"
	"fmt"
	go_grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"log"
	"net"
)

const Token = "aaa"
const Authority = "aaaa"

func main() {
	port := flag.Int("p", 1234, "listen port")
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	// 新建grpc服务,并传入拦截器进行认证检查
	grpcServer := grpc.NewServer(
		grpc.StreamInterceptor(go_grpc_middleware.ChainStreamServer(
			grpc_auth.StreamServerInterceptor(auth),
		)),
		grpc.UnaryInterceptor(go_grpc_middleware.ChainUnaryServer(
			grpc_auth.UnaryServerInterceptor(auth),
		)),
	)
	shn := &sayHelloNotify{}
	shn.last = &pb.NotifyMes{}
	shn.init()
	// 注册grpc实例
	pb.RegisterTryServiceServer(grpcServer, shn)
	log.Printf("server start")

	// 使用http进行rpc调试.浏览器访问URL http://127.0.0.1:8080/TryService/Notify 调试双向流
	// http://127.0.0.1:8080/TryService/Notify/TryService/LastNotify 调试一元调用
	go func() {
		var opts []grpc.DialOption
		opts = append(opts, grpc.WithInsecure())
		opts = append(opts, grpc.WithAuthority(Authority))
		pb.Serve(":8080", "127.0.0.1:1234", pb.DefaultHtmlStringer, opts...)
	}()
	grpcServer.Serve(lis)
}

// auth 对传入的metadata中的:authority内容进行判断,失败后再对token进行判断
func auth(ctx context.Context) (ctx1 context.Context, err error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		err = status.Error(codes.Unauthenticated, "token信息不存在")
		return
	}
	if authority, exists := md[":authority"]; exists && len(authority) > 0 {
		if authority[0] == Authority {
			ctx1 = ctx
			return
		}
	}
	tokenMD := md["token"]
	if tokenMD == nil || len(tokenMD) < 1 {
		err = status.Error(codes.Unauthenticated, "token信息不存在")
		return
	}
	if err = tokenValidate(tokenMD[0]); err != nil {
		err = status.Error(codes.Unauthenticated, err.Error())
		return
	}
	ctx1 = ctx
	return
}

func tokenValidate(token string) (err error) {
	if token != Token {
		err = fmt.Errorf("token认证错误")
		return
	}
	return
}
