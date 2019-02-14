#GRPC示例

此示例实现了
1. grpc的简单认证
2. 一元调用和服务端推送功能
2. 示例了grpc转restful

protobuf 结构如下:

    syntax = "proto3";

    package pb;
    option optimize_for = CODE_SIZE;
    import "google/protobuf/any.proto";
    import "google/protobuf/timestamp.proto";

    service TryService {
        rpc Notify (TryRequest) returns (stream NotifyMes);
        rpc LastNotify (TryRequest) returns (NotifyMes);
    }

    message TryRequest {
        string id = 1;
    }

    message NotifyMes {
        string Content = 1;
        google.protobuf.Any Details = 2;
        map<string, google.protobuf.Timestamp> UserList = 3;
    }

    message Address {
        string IP = 1;
        int32 Port = 2;
    }

server是服务端,client是客户端,gateway是grpc转restful的代码