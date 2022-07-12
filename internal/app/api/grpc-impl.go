package api

import (
	"context"
	"errors"

	"github.com/romm80/shortener.git/internal/app"
	"github.com/romm80/shortener.git/internal/app/models"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/romm80/shortener.git/internal/app/service"
	pb "github.com/romm80/shortener.git/pkg/shortener"
)

type Shortener struct {
	pb.UnimplementedShortenerServer
	Service *service.Services
}

func (s *Shortener) PingDB(ctx context.Context, in *emptypb.Empty) (*emptypb.Empty, error) {
	if err := s.Service.Storage.Ping(); err != nil {
		return &emptypb.Empty{}, status.Error(codes.Internal, "DB not found")
	}
	return &emptypb.Empty{}, nil
}

func (s *Shortener) Add(ctx context.Context, req *pb.OriginURL) (resp *pb.ShortURL, err error) {
	userID := ctx.Value("userid").(uint64)
	resp = &pb.ShortURL{}

	resp.Result, err = s.Service.Add(req.Url, userID)
	if err != nil && !errors.Is(err, app.ErrConflictURLID) {
		return resp, status.Error(codes.Internal, err.Error())
	}
	if errors.Is(err, app.ErrConflictURLID) {
		return resp, status.Error(codes.AlreadyExists, err.Error())
	}

	return
}

func (s *Shortener) Get(ctx context.Context, req *pb.RequestID) (resp *pb.OriginURL, err error) {
	resp = &pb.OriginURL{}

	resp.Url, err = s.Service.Get(req.UrlID)
	if err != nil && !errors.Is(err, app.ErrDeletedURL) && !errors.Is(err, app.ErrLinkNoFound) {
		return resp, status.Error(codes.Internal, err.Error())
	}
	if errors.Is(err, app.ErrDeletedURL) {
		return resp, status.Error(codes.DataLoss, err.Error())
	}
	if errors.Is(err, app.ErrLinkNoFound) {
		return resp, status.Error(codes.NotFound, err.Error())
	}
	return
}

func (s Shortener) BatchURLs(ctx context.Context, req *pb.RequestBatchURL) (*pb.ResponseBatchURL, error) {
	userID := ctx.Value("userid").(uint64)
	resp := &pb.ResponseBatchURL{}

	var reqBatch []models.RequestBatch
	for _, v := range req.BatchURL {
		reqBatch = append(reqBatch, models.RequestBatch{
			CorrelationID: v.CorrelationID,
			OriginalURL:   v.OriginalURL,
		})
	}

	respBatch, err := s.Service.AddBatch(reqBatch, userID)
	if err != nil && !errors.Is(err, app.ErrConflictURLID) {
		return resp, status.Error(codes.Internal, err.Error())
	}
	if errors.Is(err, app.ErrConflictURLID) {
		return resp, status.Error(codes.AlreadyExists, err.Error())
	}

	for _, v := range respBatch {
		resp.BatchURL = append(resp.BatchURL, &pb.ResponseBatchURL_URL{
			CorrelationID: v.CorrelationID,
			ShortURL:      v.ShortURL,
		})
	}

	return resp, nil
}

func (s *Shortener) GetUserURLs(ctx context.Context, in *emptypb.Empty) (*pb.UserURLs, error) {
	userID := ctx.Value("userid").(uint64)
	resp := &pb.UserURLs{}

	res, err := s.Service.GetUserURLs(userID)
	if err != nil {
		return resp, status.Error(codes.Internal, err.Error())
	}
	if len(res) == 0 {
		return resp, status.Error(codes.NotFound, "links not found")
	}

	for _, v := range res {
		resp.URLs = append(resp.URLs, &pb.UserURLs_URL{
			ShortURL:    v.ShortURL,
			OriginalURL: v.OriginalURL,
		})
	}

	return resp, nil
}

func (s *Shortener) DeleteUserURLs(ctx context.Context, req *pb.DeleteIDs) (*emptypb.Empty, error) {
	userID := ctx.Value("userid").(uint64)

	if len(req.ID) == 0 {
		return &emptypb.Empty{}, status.Error(codes.InvalidArgument, "empty link ids")
	}
	s.Service.DeleteWorker.Add(userID, req.ID)
	return &emptypb.Empty{}, nil
}

func (s *Shortener) Stats(ctx context.Context, req *emptypb.Empty) (*pb.StatsMsg, error) {
	resp := &pb.StatsMsg{}
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return resp, status.Errorf(codes.Unavailable, "")
	}

	v := md["X-Real-IP"]
	if len(v) == 0 {
		return resp, status.Errorf(codes.Unavailable, "")
	}
	res, err := s.Service.GetStats()
	if err != nil {
		return resp, status.Error(codes.Internal, err.Error())
	}
	resp.Users = int32(res.Users)
	resp.URLS = int32(res.URLs)
	return resp, nil
}
