package api

import (
	"context"

	"github.com/romm80/shortener.git/internal/app/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func (s *Shortener) AuthInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		skipAuth := map[string]struct{}{
			"/shortener.Shortener/PingDB": {},
			"/shortener.Shortener/Get":    {},
		}

		if _, inMap := skipAuth[info.FullMethod]; inMap {
			return handler(ctx, req)
		}

		var userID uint64

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.DataLoss, "failed to et metadata")
		}

		v := md["userid"]
		if len(v) == 0 || !service.ValidUserID(v[0], &userID) {
			if userID, err = s.Service.NewUser(); err != nil {
				return nil, status.Error(codes.Internal, err.Error())
			}
			signedID, err := service.SignUserID(userID)
			if err != nil {
				return nil, status.Error(codes.Internal, err.Error())
			}
			h := metadata.New(map[string]string{"userid": signedID})
			if err := grpc.SendHeader(ctx, h); err != nil {
				return nil, status.Error(codes.Internal, err.Error())
			}
		}

		return handler(context.WithValue(ctx, "userid", userID), req)
	}
}
