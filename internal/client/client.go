package client

import pb "github.com/size12/gophkeeper/protocols/grpc"

type Client struct {
	pb.GophkeeperClient
}
