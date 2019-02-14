all:
	protoc -I/usr/local/include -I. \
	-I${GOPATH}/src \
	--go_out=plugins=grpc:. \
	try.proto

	protoc -I/usr/local/include -I. \
    -I${GOPATH}/src \
    -I${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
	--grpc-gateway_out=logtostderr=true:. \
	try.proto

	protoc -I/usr/local/include -I. \
    -I${GOPATH}/src \
    -I${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
    --swagger_out=logtostderr=true:. \
    try.proto

	protoc --letmegrpc_out=. \
	*.proto