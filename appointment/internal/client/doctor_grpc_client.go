package client

import (
	"context"
	"time"

	"github.com/Akendo/assigment1/appointment/internal/usecase"
	doctorpb "github.com/Akendo/assigment1/doctor/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type DoctorGRPCClient struct {
	client  doctorpb.DoctorServiceClient
	timeout time.Duration
}

func NewDoctorGRPCClient(client doctorpb.DoctorServiceClient) *DoctorGRPCClient {
	return &DoctorGRPCClient{
		client:  client,
		timeout: 5 * time.Second,
	}
}

func (c *DoctorGRPCClient) Exists(id string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	_, err := c.client.GetDoctor(ctx, &doctorpb.GetDoctorRequest{Id: id})
	if err != nil {
		st, ok := status.FromError(err)
		if !ok {
			return false, usecase.ErrDoctorServiceUnavailable
		}

		switch st.Code() {
		case codes.NotFound:
			return false, nil
		case codes.Unavailable, codes.DeadlineExceeded:
			return false, usecase.ErrDoctorServiceUnavailable
		default:
			return false, usecase.ErrDoctorServiceUnavailable
		}
	}

	return true, nil
}
