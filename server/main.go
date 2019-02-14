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
	"html"
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
	pb.RegisterTryServiceServer(grpcServer, shn)
	log.Printf("server start")

	// 浏览器访问URL /TryService/Notify 和 /TryService/LastNotify
	go func() {
		var opts []grpc.DialOption
		opts = append(opts, grpc.WithInsecure())
		opts = append(opts, grpc.WithAuthority(Authority))
		pb.Serve(":8080", "127.0.0.1:1234", pb.DefaultHtmlStringer, opts...)
	}()
	grpcServer.Serve(lis)
}

var golangStructHtmlStringer = func(req, resp interface{}) ([]byte, error) {
	header := []byte("<p><div class=\"container\"><pre>")
	data := []byte(html.EscapeString(fmt.Sprintf("%+v", resp)))
	footer := []byte("</pre></div></p>")
	return append(append(header, data...), footer...), nil
}

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
