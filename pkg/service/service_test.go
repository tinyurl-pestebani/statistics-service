package service

import (
	"context"
	"fmt"
	pb "github.com/tinyurl-pestebani/go-proto-pkg/pkg/pb/v1"
	"github.com/tinyurl-pestebani/statistics-database/pkg/db/mockdb"
	"github.com/tinyurl-pestebani/statistics-service/pkg/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"io"
	"log"
	"net"
	"testing"
	"time"
)

var (
	tag = "123456ab"
)

func TestMain(m *testing.M) {
	dbSTats, err := mockdb.NewMockDB()

	ctx := context.Background()

	dbSTats.InsertStatisticPoint(ctx, tag, time.Date(2025, time.August, 1, 1, 2, 3, 4, time.UTC))
	dbSTats.InsertStatisticPoint(ctx, tag, time.Date(2025, time.August, 1, 1, 3, 3, 4, time.UTC))
	dbSTats.InsertStatisticPoint(ctx, tag, time.Date(2025, time.August, 1, 1, 4, 3, 4, time.UTC))
	dbSTats.InsertStatisticPoint(ctx, tag, time.Date(2025, time.August, 1, 2, 2, 3, 4, time.UTC))
	dbSTats.InsertStatisticPoint(ctx, tag, time.Date(2025, time.August, 1, 3, 2, 3, 4, time.UTC))
	dbSTats.InsertStatisticPoint(ctx, tag, time.Date(2025, time.August, 1, 6, 2, 3, 4, time.UTC))
	dbSTats.InsertStatisticPoint(ctx, tag, time.Date(2025, time.August, 1, 9, 2, 3, 4, time.UTC))
	dbSTats.InsertStatisticPoint(ctx, tag, time.Date(2025, time.August, 1, 10, 2, 3, 4, time.UTC))
	dbSTats.InsertStatisticPoint(ctx, tag, time.Date(2025, time.August, 1, 11, 2, 30, 4, time.UTC))

	if err != nil {
		panic(err)
	}

	svc, err := NewStatisticsService(dbSTats)

	if err != nil {
		panic(fmt.Errorf("NewStatisticsService() error = %v, wantErr %v", err, nil))
	}

	conf, err := config.NewServiceConfigFromEnv()

	if err != nil {
		panic(fmt.Errorf("NewServiceConfigFromEnv() error = %v, wantErr %v", err, nil))
	}

	dir := fmt.Sprintf("[::]:%d", conf.Port)

	lis, err := net.Listen("tcp", dir)

	if err != nil {
		panic(fmt.Errorf("net.Listen() error = %v, wantErr %v", err, nil))
	}

	grpcServer := grpc.NewServer()

	pb.RegisterStatisticsServiceServer(grpcServer, svc)

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	defer grpcServer.GracefulStop()
}

func TestStatisticsService_GetStatistics(t *testing.T) {
	ctx := context.Background()

	conf, err := config.NewServiceConfigFromEnv()

	if err != nil {
		t.Fatalf("NewServiceConfigFromEnv() error = %v, wantErr %v", err, nil)
	}

	dir := fmt.Sprintf("[::]:%d", conf.Port)

	conn, err := grpc.NewClient(dir, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		t.Fatalf("NewClient() error = %v, wantErr %v", err, nil)
	}

	client := pb.NewStatisticsServiceClient(conn)

	defer conn.Close()

	start := time.Date(2025, time.August, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2025, time.August, 2, 0, 0, 0, 0, time.UTC)

	tz := "UTC"

	req := &pb.GetStatisticsRequest{
		Tag:              tag,
		IntervalStart:    &timestamppb.Timestamp{Seconds: start.Unix()},
		IntervalEnd:      &timestamppb.Timestamp{Seconds: end.Unix()},
		IntervalDuration: &durationpb.Duration{Seconds: 3600}, // 1 hour
		Timezone:         &tz,
	}

	stream, err := client.GetStatistics(ctx, req)

	if err != nil {
		t.Errorf("GetStatistics() error = %v, wantErr %v", err, nil)
	}

	var resps []*pb.GetStatisticsResponse

	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			t.Errorf("GetStatistics() error = %v, wantErr %v", err, nil)
		}

		resps = append(resps, resp)
	}

	if len(resps) != 25 {
		t.Errorf("GetStatistics() error = %v, wantErr %v", err, nil)
	}

	stats, ok := resps[0].StatisticsResponseOneOf.(*pb.GetStatisticsResponse_Stats)

	if !ok {
		t.Errorf("GetStatistics() error = %v, wantErr %v", err, nil)
	}

	if stats.Stats.GetTotalPoints() != 24 {
		t.Errorf("GetStatistics() error = %v, wantErr %v", err, nil)
	}

	point1, ok := resps[2].StatisticsResponseOneOf.(*pb.GetStatisticsResponse_StatisticPoint)

	if !ok {
		t.Errorf("GetStatistics() error = %v, wantErr %v", err, nil)
	}

	if point1.StatisticPoint.GetTotalVisits() != 3 {
		t.Errorf("GetStatistics() error = %v, wantErr %v", err, nil)
	}
}

func TestStatisticsService_Ping(t *testing.T) {
	ctx := context.Background()
	conf, err := config.NewServiceConfigFromEnv()
	if err != nil {
		t.Fatalf("NewServiceConfigFromEnv() error = %v, wantErr %v", err, nil)
	}

	dir := fmt.Sprintf("[::]:%d", conf.Port)
	conn, err := grpc.NewClient(dir, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		t.Fatalf("NewClient() error = %v, wantErr %v", err, nil)
	}
	client := pb.NewStatisticsServiceClient(conn)

	defer conn.Close()

	resp, err := client.Ping(ctx, &pb.PingRequest{})

	if err != nil {
		t.Errorf("Ping() error = %v, wantErr %v", err, nil)
	}

	if resp.GetResponse() != "pong" {
		t.Errorf("Ping() error = %v, wantErr %v", err, nil)
	}
}
