package main

import (
	"context"
	"fmt"
	"log"
	"net"

	pb "CryptoParser/proto"

	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedCryptoServer
}

func (s *server) CryptoPrice(ctx context.Context, req *pb.PriceRequest) (*pb.PriceResponse, error) {
	price, err := RequestTicker(req.Name)
	if err != nil {
		return nil, fmt.Errorf("Не удалось получить цену для %s", req.Name)
	}
	return &pb.PriceResponse{
		Message: fmt.Sprintf("Цена %s: $%s", req.Name, price),
	}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterCryptoServer(s, &server{})
	log.Println("Server started on :50052")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
