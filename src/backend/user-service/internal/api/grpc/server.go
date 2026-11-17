package grpc

import (
	pb "proto/user"
)

type Server struct {
	pb.UnimplementedUserServer
}

func NewServer() *Server {
	return &Server{}
}
