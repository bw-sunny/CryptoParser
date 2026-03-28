package main

import (
	"context"
	"fmt"
	"log"
	"net"

	pb "CryptoParser/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
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

func (s *server) CryptoRatio(ctx context.Context, req *pb.RatioRequest) (*pb.RatioResponse, error) {
	buyRatio, sellRatio, longShortRatio, err := RequestRatio(req.Name)
	if err != nil {
		return nil, fmt.Errorf("Не удалось получить данные long/short ratio для %s", req.Name)
	}

	return &pb.RatioResponse{
		Message: fmt.Sprintf("Long/Short Ratio %s: Buy=%s%%, Sell=%s%%, L/S=%s",
			req.Name, buyRatio, sellRatio, longShortRatio),
		BuyRatio:       buyRatio,
		SellRatio:      sellRatio,
		LongShortRatio: longShortRatio,
	}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterCryptoServer(s, &server{})
	reflection.Register(s)
	log.Println("Server started on :50052")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
