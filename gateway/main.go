package main

import (
	"context"
	pb "../pb"
	"flag"
	"log"
	"net/http"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
	"strconv"
)

var (
	port = flag.Int("p", 1234, "listen port")
)

const Authority = "aaaa"

func run() error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	gwmux := runtime.NewServeMux()
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	opts = append(opts, grpc.WithAuthority("aaaa"))
	err := pb.RegisterTryServiceHandlerFromEndpoint(ctx, gwmux, "127.0.0.1:"+strconv.Itoa(*port), opts)
	if err != nil {
		return err
	}
	mux:=http.NewServeMux()
	mux.Handle("/",gwmux)
	mux.Handle("/swagger/", http.FileServer(http.Dir(".")))
	log.Printf("Listen :8080")
	return http.ListenAndServe(":8080", mux)
}

func main() {
	flag.Parse()
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
