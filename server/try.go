package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
	. "../pb"
	uuid "github.com/google/uuid"
	"github.com/golang/protobuf/ptypes"
	timestamp "github.com/golang/protobuf/ptypes/timestamp"
	"google.golang.org/grpc/peer"
)

type sayHelloNotify struct {
	last    *NotifyMes
	count   int
	msgChan map[string]chan string
	list    map[string]*timestamp.Timestamp
	m       sync.Mutex
}

// 启动消息发生器
func (self *sayHelloNotify) init() {
	self.msgChan = make(map[string]chan string)
	self.list = make(map[string]*timestamp.Timestamp)
	go func() {
		timer := time.Tick(500 * time.Millisecond)
		for {
			<-timer
			self.m.Lock()
			if len(self.msgChan) == 0 {
				self.m.Unlock()
				continue
			}
			self.count++
			log.Printf("%d", self.count)
			msg := strconv.Itoa(self.count)
			self.m.Unlock()
			for id, s := range self.msgChan {
				select {
				case s <- msg:
				default:
					log.Printf("put msg to chan error:id=%s,count=%s", id, s)
				}
			}
		}
	}()
}

// LastNotify 是一元调用模式
func (self *sayHelloNotify) LastNotify(context.Context, *TryRequest) (*NotifyMes, error) {
	if self.last == nil {
		return nil, fmt.Errorf("message is nil")
	}
	return self.last, nil
}

// Notify 是双向流模式
func (s *sayHelloNotify) Notify(r *TryRequest, stream TryService_NotifyServer) error {
	c := make(chan string)
	u := uuid.New()
	s.m.Lock()
	key := r.GetId() + ":" + u.String()
	s.msgChan[key] = c
	s.list[key] = &timestamp.Timestamp{Seconds: time.Now().Unix()}
	s.m.Unlock()
	defer func() {
		s.m.Lock()
		delete(s.list, key)
		delete(s.msgChan, key)
		if len(s.msgChan) == 0 {
			s.count = 0
		}
		s.m.Unlock()
	}()
	timeout := time.Tick(time.Hour)
	for {
		select {
		// 超时返回,服务端会自动断开流
		case <-timeout:
			return fmt.Errorf("超时")
			// ctx被取消(流断开)
		case <-stream.Context().Done():
			log.Printf("client down(%s):%v", key, stream.Context().Err())
			return fmt.Errorf("client down error")
		case m := <-c:
			hello := "hello " + r.GetId() + "." + m
			ip, portStr, _ := getClientIP(stream.Context())
			port, _ := strconv.Atoi(portStr)
			// 序列化Any
			any, _ := ptypes.MarshalAny(&Address{IP: ip, Port: int32(port)})
			msg := NotifyMes{Content: hello, UserList: s.list, Details: any}
			s.last = &msg
			// 发送消息
			err := stream.Send(&msg)
			if err != nil {
				log.Printf(err.Error())
			}
		}
	}
	return nil
}

// getClientIP 通过节点信息获取地址
func getClientIP(ctx context.Context) (ip, port string, err error) {
	// 获取节点信息
	pr, ok := peer.FromContext(ctx)
	if !ok {
		err = fmt.Errorf("[getClinetIP] invoke FromContext() failed")
		return
	}
	if pr.Addr == net.Addr(nil) {
		err = fmt.Errorf("[getClientIP] peer.Addr is nil")
		return
	}
	// 地址格式为 ip:port
	addSlice := strings.Split(pr.Addr.String(), ":")
	ip = addSlice[0]
	port = addSlice[1]
	return
}
