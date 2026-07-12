package server

import (
	"context"
	"time"

	pb "github.com/anvitha0403/golog/internal/server/api/v1"
	"github.com/anvitha0403/golog/internal/server/auth"
	"github.com/anvitha0403/golog/internal/server/kvstore"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"

	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"go.opencensus.io/plugin/ocgrpc"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

// END: config
// START: config
type Config struct {
	Store       kvstore.Store
	Authorizer  *auth.Authorizer
	GetServerer GetServerer
}

type GetServerer interface {
	GetServers() ([]*pb.Server, error)
}

const (
	objectWildcard = "*"
)

type grpcServer struct {
	*Config
	pb.UnimplementedKVStoreServiceServer
}

func newgrpcServer(config *Config) (*grpcServer, error) {

	return &grpcServer{
		Config: &Config{
			Store:      config.Store,
			Authorizer: &auth.Authorizer{},
		},
	}, nil

}

func (s *grpcServer) Put(ctx context.Context, pair *pb.KVPair) (*pb.Response, error) {
	if err := s.Authorizer.Authorize(
		subject(ctx),
		objectWildcard,
		kvstore.OPERATION_PUT,
	); err != nil {
		return nil, err
	}
	op, err := s.Store.Put(pair.Key, pair.Value)
	if err != nil {

		return nil, err
	}

	return &pb.Response{Value: op}, nil
}

func (s *grpcServer) Get(ctx context.Context, pair *pb.KVPair) (*pb.Response, error) {
	if err := s.Authorizer.Authorize(
		subject(ctx),
		objectWildcard,
		kvstore.OPERATION_GET,
	); err != nil {
		return nil, err
	}
	value, err := s.Store.Get(pair.Key)
	if err != nil {
		if err == kvstore.ErrKeyDoesntExist {
			return nil, &pb.ErrKeyDoesntExist{}
		}
		return nil, err
	}
	return &pb.Response{Value: value}, nil
}

func (s *grpcServer) Delete(ctx context.Context, pair *pb.KVPair) (*pb.Empty, error) {
	if err := s.Authorizer.Authorize(
		subject(ctx),
		objectWildcard,
		kvstore.OPERATION_DEL,
	); err != nil {
		return nil, err
	}
	err := s.Store.Del(pair.Key)
	if err != nil {
		if err == kvstore.ErrKeyDoesntExist {
			return nil, &pb.ErrKeyDoesntExist{}
		}
		return nil, err
	}

	return &pb.Empty{}, nil
}

// START: logger
func NewGRPCServer(config *Config, grpcOpts ...grpc.ServerOption) (
	*grpc.Server,
	error,
) {
	logger := zap.L().Named("server")
	zapOpts := []grpc_zap.Option{
		grpc_zap.WithDurationField(
			func(duration time.Duration) zapcore.Field {
				return zap.Int64(
					"grpc.time_ns",
					duration.Nanoseconds(),
				)
			},
		),
	}

	// START: metrics_traces
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})
	err := view.Register(ocgrpc.DefaultServerViews...)
	if err != nil {
		return nil, err
	}
	// END: metrics_traces

	// START: grpc_opts
	grpcOpts = append(grpcOpts,
		grpc.StreamInterceptor(
			grpc_middleware.ChainStreamServer(
				// START_HIGHLIGHT
				grpc_ctxtags.StreamServerInterceptor(),
				grpc_zap.StreamServerInterceptor(logger, zapOpts...),
				// END_HIGHLIGHT
				grpc_auth.StreamServerInterceptor(authenticate),
			)), grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			// START_HIGHLIGHT
			grpc_ctxtags.UnaryServerInterceptor(),
			grpc_zap.UnaryServerInterceptor(logger, zapOpts...),
			// END_HIGHLIGHT
			grpc_auth.UnaryServerInterceptor(authenticate),
		)),
		// START_HIGHLIGHT
		grpc.StatsHandler(&ocgrpc.ServerHandler{}),
		// END_HIGHLIGHT
	)
	// END: grpc_opts
	gsrv := grpc.NewServer(grpcOpts...)
	srv, err := newgrpcServer(config)
	if err != nil {
		return nil, err
	}
	pb.RegisterKVStoreServiceServer(gsrv, srv)
	return gsrv, nil
}

// START: authorizer
type Authorizer interface {
	Authorize(subject, object, action string) error
}

// END: authorizer

// START: authenticate
func authenticate(ctx context.Context) (context.Context, error) {
	peer, ok := peer.FromContext(ctx)
	if !ok {
		return ctx, status.New(
			codes.Unknown,
			"couldn't find peer info",
		).Err()
	}

	if peer.AuthInfo == nil {
		return ctx, status.New(
			codes.Unauthenticated,
			"no transport security being used",
		).Err()
	}

	tlsInfo := peer.AuthInfo.(credentials.TLSInfo)
	subject := tlsInfo.State.VerifiedChains[0][0].Subject.CommonName
	ctx = context.WithValue(ctx, subjectContextKey{}, subject)

	return ctx, nil
}

func subject(ctx context.Context) string {
	return ctx.Value(subjectContextKey{}).(string)
}

type subjectContextKey struct{}

// END: authenticate
