//go:generate protoc --proto_path=../proto --go_out=../proto --go_opt=paths=source_relative --go-grpc_out=../proto --go-grpc_opt=paths=source_relative ../proto/stonks.proto
//go:generate python3 -m grpc_tools.protoc --proto_path=../proto --python_out=../proto --grpc_python_out=../proto ../proto/stonks.proto

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/golang/protobuf/ptypes/empty"
	_ "github.com/piquette/finance-go"
	pb "github.com/team-stonks/stonks-api/proto"
	"google.golang.org/grpc"
	tb "gopkg.in/tucnak/telebot.v2"
)

const (
	port = ":50051"
)

type server struct {
	pb.UnimplementedStonksApiServer
	bot *tb.Bot
}

type FinancialStat struct {
	Value float64 `json:"raw"`
}

type FinancialData struct {
	EbitdaMargins           FinancialStat `json:"ebitdaMargins"`
	ProfitMargins           FinancialStat `json:"profitMargins"`
	GrossMargins            FinancialStat `json:"grossMargins"`
	OperatingCashflow       FinancialStat `json:"operatingCashflow"`
	RevenueGrowth           FinancialStat `json:"revenueGrowth"`
	OperatingMargins        FinancialStat `json:"operatingMargins"`
	Ebitda                  FinancialStat `json:"ebitda"`
	TargetLowPrice          FinancialStat `json:"targetLowPrice"`
	GrossProfits            FinancialStat `json:"grossProfits"`
	FreeCashflow            FinancialStat `json:"freeCashflow"`
	TargetMedianPrice       FinancialStat `json:"targetMedianPrice"`
	CurrentPrice            FinancialStat `json:"currentPrice"`
	EarningsGrowth          FinancialStat `json:"earningsGrowth"`
	CurrentRatio            FinancialStat `json:"currentRatio"`
	ReturnOnAssets          FinancialStat `json:"returnOnAssets"`
	NumberOfAnalystOpinions FinancialStat `json:"numberOfAnalystOpinions"`
	TargetMeanPrice         FinancialStat `json:"targetMeanPrice"`
	DebtToEquity            FinancialStat `json:"debtToEquity"`
	ReturnOnEquity          FinancialStat `json:"returnOnEquity"`
	TargetHighPrice         FinancialStat `json:"targetHighPrice"`
	TotalCash               FinancialStat `json:"totalCash"`
	TotalDebt               FinancialStat `json:"totalDebt"`
	TotalRevenue            FinancialStat `json:"totalRevenue"`
	TotalCashPerShare       FinancialStat `json:"totalCashPerShare"`
}

type StatisticResponse struct {
	FinancialData FinancialData `json:"financialData"`
}

func (s *server) TelegramNotification(_ context.Context, r *pb.TelegramRequest) (*empty.Empty, error) {
	log.Println("Called TelegramNotify")
	s.bot.Send(tb.ChatID(266737912), r.Message)
	return &empty.Empty{}, nil
}

func getCompanyStatsImpl(symbol string) (*StatisticResponse, error) {
	url := fmt.Sprintf("https://apidojo-yahoo-finance-v1.p.rapidapi.com/stock/v2/get-summary?symbol=%s&region=US", symbol)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("x-rapidapi-key", os.Getenv("YAHOO_TOKEN"))
	req.Header.Add("x-rapidapi-host", "apidojo-yahoo-finance-v1.p.rapidapi.com")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	resp := StatisticResponse{}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (s *server) GetCompanyStats(_ context.Context, r *pb.CompanyStatsRequest) (*pb.CompanyStatsResponse, error) {
	resp, err := getCompanyStatsImpl(r.Figi)
	if err != nil {
		return nil, err
	}
	return &pb.CompanyStatsResponse{
		EbitdaMargins:           resp.FinancialData.EbitdaMargins.Value,
		ProfitMargins:           resp.FinancialData.ProfitMargins.Value,
		GrossMargins:            resp.FinancialData.GrossMargins.Value,
		OperatingCashflow:       resp.FinancialData.OperatingCashflow.Value,
		RevenueGrowth:           resp.FinancialData.RevenueGrowth.Value,
		OperatingMargins:        resp.FinancialData.OperatingMargins.Value,
		Ebitda:                  resp.FinancialData.Ebitda.Value,
		TargetLowPrice:          resp.FinancialData.TargetLowPrice.Value,
		GrossProfits:            resp.FinancialData.GrossProfits.Value,
		FreeCashflow:            resp.FinancialData.FreeCashflow.Value,
		TargetMedianPrice:       resp.FinancialData.TargetMedianPrice.Value,
		CurrentPrice:            resp.FinancialData.CurrentPrice.Value,
		EarningsGrowth:          resp.FinancialData.EarningsGrowth.Value,
		CurrentRatio:            resp.FinancialData.CurrentRatio.Value,
		ReturnOnAssets:          resp.FinancialData.ReturnOnAssets.Value,
		NumberOfAnalystOpinions: resp.FinancialData.NumberOfAnalystOpinions.Value,
		TargetMeanPrice:         resp.FinancialData.TargetMeanPrice.Value,
		DebtToEquity:            resp.FinancialData.DebtToEquity.Value,
		ReturnOnEquity:          resp.FinancialData.ReturnOnEquity.Value,
		TargetHighPrice:         resp.FinancialData.TargetHighPrice.Value,
		TotalCash:               resp.FinancialData.TotalCash.Value,
		TotalDebt:               resp.FinancialData.TotalDebt.Value,
		TotalRevenue:            resp.FinancialData.TotalRevenue.Value,
		TotalCashPerShare:       resp.FinancialData.TotalCashPerShare.Value,
	}, nil
}

func main() {
	log.Println("Starting server")

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	const telegramTokenEnv string = "TELEGRAM_BOT_TOKEN"
	b, bot_err := tb.NewBot(tb.Settings{Token: os.Getenv(telegramTokenEnv),
		Poller: &tb.LongPoller{Timeout: 10 * time.Second}})

	if bot_err != nil {
		log.Fatal(bot_err)
		return
	}

	b.Handle("/hello", func(m *tb.Message) {
		b.Send(m.Sender, "privet kozel")
	})

	go b.Start()

	pb.RegisterStonksApiServer(s, &server{bot: b})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
