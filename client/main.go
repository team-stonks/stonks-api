package main

import (
	"context"
	"fmt"
	"github.com/golang/protobuf/ptypes"
	"log"
	"time"

	pb "github.com/team-stonks/stonks-api/proto"
	"google.golang.org/grpc"
)

const (
	address = "localhost:50051"
)

func main() {
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewStonksApiClient(conn)
	from := ptypes.TimestampNow()
	to := time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)
	toProto, err := ptypes.TimestampProto(to)
	if err != nil {
		fmt.Println(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	resp, err := c.GetHistory(ctx, &pb.HistoryRequest{Figi: "INTC", From: from, To: toProto, Interval: "week"})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}

	fmt.Println(len(resp.Candles))
}
