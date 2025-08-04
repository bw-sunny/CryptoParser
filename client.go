// package main

// import (
// 	"context"
// 	"fmt"
// 	"log"
// 	"time"

// 	pb "CryptoParser/proto"

// 	"google.golang.org/grpc"
// 	"google.golang.org/grpc/credentials/insecure"
// )

// func main() {
// 	conn, err := grpc.Dial("localhost:50052", grpc.WithTransportCredentials(insecure.NewCredentials()))
// 	if err != nil {
// 		log.Fatalf("did not connect: %v", err)
// 	}
// 	defer conn.Close()

// 	c := pb.NewCryptoClient(conn)
// 	name := "BTC" // по умолчанию BTC
// 	fmt.Printf("Введите название тикера: \n")
// 	fmt.Scan(&name)

// 	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
// 	defer cancel()

// 	r, err := c.CryptoPrice(ctx, &pb.PriceRequest{Name: name})
// 	if err != nil {
// 		log.Fatalf("could not greet: %v", err)
// 	}
// 	log.Printf("%s", r.Message)
// }
