package service

import (
	"github.com/tinyurl-pestebani/statistics-database/pkg/db"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// convertStatDatabaseError converts a database error to a gRPC status error.
func convertStatDatabaseError(err error) error {
	switch err.(type) {
	case nil:
		return status.Error(codes.OK, "")
	case *db.InvalidDataError:
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}
