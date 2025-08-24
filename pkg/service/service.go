package service

import (
	"context"
	pb "github.com/tinyurl-pestebani/go-proto-pkg/pkg/pb/v1"
	"github.com/tinyurl-pestebani/statistics-database/pkg/db"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	otlpcode "go.opentelemetry.io/otel/codes"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

const (
	name = "statistics-service"
)

var (
	tracer = otel.Tracer(name)
	logger = otelslog.NewLogger(name)
)

// StatisticsService is the implementation of the StatisticsServiceServer interface.
type StatisticsService struct {
	pb.UnimplementedStatisticsServiceServer
	statsDB db.StatisticsDatabase
}

// NewStatisticsService creates a new StatisticsService.
func NewStatisticsService(statsDB db.StatisticsDatabase) (*StatisticsService, error) {
	return &StatisticsService{statsDB: statsDB}, nil
}

// Close closes the database connection.
func (ss *StatisticsService) Close() error {
	return ss.statsDB.Close()
}

// Ping implementation of StatisticsServiceServer interface.
func (ss *StatisticsService) Ping(ctx context.Context, _ *pb.PingRequest) (*pb.PingResponse, error) {
	ctx, span := tracer.Start(ctx, "Ping")
	defer span.End()

	logger.InfoContext(ctx, "Ping function", "result", "Pong")
	span.SetStatus(otlpcode.Ok, "Executed successfully")

	return &pb.PingResponse{
		Response: "pong",
	}, status.Error(codes.OK, "")
}

// GetStatistics implementation of StatisticsServiceServer interface.
func (ss *StatisticsService) GetStatistics(in *pb.GetStatisticsRequest, stream pb.StatisticsService_GetStatisticsServer) error {
	ctx := stream.Context()
	ctx, span := tracer.Start(ctx, "GetStatistics")
	defer span.End()

	urlId := in.GetTag()
	start := in.GetIntervalStart().AsTime()
	end := in.GetIntervalEnd().AsTime()
	dur := in.GetIntervalDuration().AsDuration()
	tz := in.GetTimezone()

	logger.InfoContext(ctx, "GetStatistics", "urlId", urlId, "start", start, "end", end, "duration", dur, "tz", tz)

	span.SetAttributes(attribute.String("url_id", urlId))
	span.SetAttributes(attribute.String("start", start.Format(time.RFC3339)))
	span.SetAttributes(attribute.String("end", end.Format(time.RFC3339)))
	span.SetAttributes(attribute.String("duration", dur.String()))
	span.SetAttributes(attribute.String("tz", tz))

	if tz == "" {
		tz = "UTC"
	}

	points, err := ss.statsDB.GetStatisticPoints(ctx, urlId, start, end, dur, tz)

	if err != nil {
		logger.ErrorContext(ctx, "GetStatistics", "err", err)
		span.RecordError(err)
		return convertStatDatabaseError(err)
	}

	totalPoints := uint32(len(points))

	if err := stream.Send(&pb.GetStatisticsResponse{
		StatisticsResponseOneOf: &pb.GetStatisticsResponse_Stats{
			Stats: &pb.GetStatisticsStats{
				TotalPoints: totalPoints,
			},
		},
	}); err != nil {
		logger.ErrorContext(ctx, "GetStatistics", "err", err)
		span.RecordError(err)
		return status.Error(codes.Internal, err.Error())
	}

	for _, point := range points {
		if err := stream.Send(&pb.GetStatisticsResponse{
			StatisticsResponseOneOf: &pb.GetStatisticsResponse_StatisticPoint{
				StatisticPoint: point.Encode(),
			},
		}); err != nil {
			logger.ErrorContext(ctx, "GetStatistics", "err", err)
			span.RecordError(err)
			return status.Error(codes.Internal, err.Error())
		}
	}

	return status.Error(codes.OK, "")
}
