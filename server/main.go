package main

import (
	"context"
	"flag"
	"fmt"
	sdk "github.com/TinkoffCreditSystems/invest-openapi-go-sdk"
	"github.com/golang/protobuf/ptypes"
	pb "github.com/team-stonks/stonks-api/proto"
	"google.golang.org/grpc"
	"log"
	"net"
	"time"
)

const (
	port = ":50051"
)

type server struct {
	pb.UnimplementedStonksApiServer
	client  *sdk.SandboxRestClient
	account *sdk.Account
}

func (s *server) GetHistory(ctx context.Context, in *pb.HistoryRequest) (*pb.HistoryResponse, error) {
	from, err := ptypes.Timestamp(in.From)
	if err != nil {
		return nil, err
	}
	to, err := ptypes.Timestamp(in.To)
	if err != nil {
		return nil, err
	}

	candles, err := s.client.Candles(ctx, from, to, sdk.CandleInterval(in.Interval), in.Figi)
	candlesProto := make([]*pb.Candle, len(candles))
	for i, c := range candles {
		ts, err := ptypes.TimestampProto(c.TS)
		if err != nil {
			return nil, err
		}
		candlesProto[i] = &pb.Candle{ClosePrice: c.ClosePrice, OpenPrice: c.OpenPrice, HighPrice: c.HighPrice, LowPrice: c.LowPrice, Volume: c.Volume, Time: ts}
	}
	return &pb.HistoryResponse{Candles: candlesProto}, nil
}

func main() {
	token := ""
	flag.StringVar(&token, "token", "", "token")
	flag.Parse()
	fmt.Println(token)
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	tinkoffClient := sdk.NewSandboxRestClient(token)
	a := initTinkoffAcc(tinkoffClient)
	pb.RegisterStonksApiServer(s, &server{client: tinkoffClient, account: &a})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func initTinkoffAcc(client *sdk.SandboxRestClient) sdk.Account {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	acc, err := client.Register(ctx, sdk.AccountTinkoff)
	if err != nil {
		log.Fatalln(err)
	}
	return acc
}
