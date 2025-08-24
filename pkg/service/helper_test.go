package service

import (
	"errors"
	"fmt"
	"github.com/tinyurl-pestebani/statistics-database/pkg/db"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"testing"
)

func TestConvertStatDatabaseError(t *testing.T) {
	err := convertStatDatabaseError(nil)

	if !errors.Is(err, status.Error(codes.OK, "")) {
		t.Errorf("ConvertStatDatabaseError() error = %v, wantErr %v", err, status.Error(codes.OK, ""))
	}

	errInvalid := db.NewInvalidDataError("invalid data")

	err = convertStatDatabaseError(errInvalid)

	if !errors.Is(err, status.Error(codes.InvalidArgument, errInvalid.Error())) {
		t.Errorf("ConvertStatDatabaseError() error = %v, wantErr %v", err, status.Error(codes.InvalidArgument, ""))
	}

	errInternal := fmt.Errorf("internal error")

	err = convertStatDatabaseError(errInternal)

	if !errors.Is(err, status.Error(codes.Internal, errInternal.Error())) {
		t.Errorf("ConvertStatDatabaseError() error = %v, wantErr %v", err, status.Error(codes.Internal, ""))
	}
}
