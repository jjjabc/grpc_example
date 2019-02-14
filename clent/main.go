package main

import (
	"../pb"
	"context"
	"flag"
	"github.com/golang/protobuf/ptypes"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"log"
)

const (
	currentToken = "aaa"
	wrongToken   = "aaa1"
)

type Auth struct {
	AppKey    string
	AppSecret string
}

func (a *Auth) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{"app_key": a.AppKey, "app_secret": a.AppSecret}, nil
}

func (a *Auth) RequireTransportSecurity() bool {
	return false
}

func main() {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	//opts = append(opts, grpc.WithAuthority("aaaa"))
	//opts = append(opts, grpc.WithPerRPCCredentials(&auth))
	target := flag.String("t", "127.0.0.1:1234", "server address(ip:port)")
	conn, err := grpc.Dial(*target, opts...)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	client := pb.NewTryServiceClient(conn)
	authStreamMd := metadata.Pairs("token", currentToken)
	authUnaryMd := metadata.Pairs("token", wrongToken)

	ctx := metadata.NewOutgoingContext(context.Background(), authStreamMd)
	nc, err := client.Notify(ctx, &pb.TryRequest{Id: "xunj"})
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		for {
			msg, err := nc.Recv()
			if err != nil {
				log.Fatal(err)
			}
			addr := pb.Address{}
			err = ptypes.UnmarshalAny(msg.Details, &addr)
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("Recv:%s\tDetail:%s:%d", msg.Content, addr.GetIP(), addr.GetPort())
		}
	}()
	ctx1 := metadata.NewOutgoingContext(context.Background(), authUnaryMd)
	last, err := client.LastNotify(ctx1, &pb.TryRequest{Id: "xunj"})
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Last:" + last.Content)
	<-make(chan bool)
}
