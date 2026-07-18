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
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

// END: config
// START: config
type Config struct {
	Store      kvstore.IDistributedStore
	Authorizer *auth.Authorizer
}
const (
	OPERATION_GETSERVERS = "GETSERVERS"
)

const (
	objectWildcard = "*"
)

type grpcServer struct {
	*Config
	pb.UnimplementedKVStoreServiceServer
}

func newgrpcServer(config *Config) (*grpcServer, error) {
	return &grpcServer{
		Config: config,
	}, nil
}

// GetServers implements [v1.KVStoreServiceServer].
func (s *grpcServer) GetServers(ctx context.Context, req *pb.GetServersRequest) (*pb.GetServersResponse, error) {
	logger := grpc_zap.Extract(ctx)
	logger.Info("Handling GetServers request")
	if err := s.Authorizer.Authorize(
		subject(ctx),
		objectWildcard,
		OPERATION_GETSERVERS,
	); err != nil {
		logger.Warn("Authorization failed for GetServers",
			zap.String("subject", subject(ctx)),
			zap.Error(err),
		)
		return nil, err
	}
	servers, err := s.Store.GetServers()
	if err != nil {
		logger.Warn("getServers failed for ",
			zap.String("subject", subject(ctx)),
			zap.Error(err),
		)
	}

	var pbservers []*pb.Server
	for _, server := range servers {
		pbservers = append(pbservers, &pb.Server{
			Id:       string(server.Id),
			RpcAddr:  string(server.RpcAddr),
			IsLeader: server.IsLeader,
		})
	}
	return &pb.GetServersResponse{Servers: pbservers}, nil

}

func (s *grpcServer) Put(ctx context.Context, pair *pb.KVPair) (*pb.Response, error) {
	logger := grpc_zap.Extract(ctx)
	logger.Info("Handling PUT request",
		zap.String("key", pair.Key),
		zap.String("value", pair.Value),
	)

	if err := s.Authorizer.Authorize(
		subject(ctx),
		objectWildcard,
		kvstore.OPERATION_PUT,
	); err != nil {
		logger.Warn("Authorization failed for PUT",
			zap.String("subject", subject(ctx)),
			zap.Error(err),
		)
		return nil, err
	}

	op, err := s.Store.Put(pair.Key, pair.Value)
	if err != nil {
		logger.Error("Store PUT failed",
			zap.String("key", pair.Key),
			zap.Error(err),
		)
		return nil, err
	}

	logger.Info("PUT successful",
		zap.String("key", pair.Key),
		zap.String("result", op),
	)
	return &pb.Response{Value: op}, nil
}

func (s *grpcServer) Get(ctx context.Context, pair *pb.KVPair) (*pb.Response, error) {
	logger := grpc_zap.Extract(ctx)
	logger.Info("Handling GET request",
		zap.String("key", pair.Key),
	)

	if err := s.Authorizer.Authorize(
		subject(ctx),
		objectWildcard,
		kvstore.OPERATION_GET,
	); err != nil {
		logger.Warn("Authorization failed for GET",
			zap.String("subject", subject(ctx)),
			zap.Error(err),
		)
		return nil, err
	}

	value, err := s.Store.Get(pair.Key)
	if err != nil {
		if err == kvstore.ErrKeyDoesntExist {
			logger.Warn("GET failed: key does not exist",
				zap.String("key", pair.Key),
			)
			return nil, &pb.ErrKeyDoesntExist{}
		}
		logger.Error("Store GET failed",
			zap.String("key", pair.Key),
			zap.Error(err),
		)
		return nil, err
	}

	logger.Info("GET successful",
		zap.String("key", pair.Key),
		zap.String("value", value),
	)
	return &pb.Response{Value: value}, nil
}

func (s *grpcServer) Delete(ctx context.Context, pair *pb.KVPair) (*pb.Empty, error) {
	logger := grpc_zap.Extract(ctx)
	logger.Info("Handling DELETE request",
		zap.String("key", pair.Key),
	)

	if err := s.Authorizer.Authorize(
		subject(ctx),
		objectWildcard,
		kvstore.OPERATION_DEL,
	); err != nil {
		logger.Warn("Authorization failed for DELETE",
			zap.String("subject", subject(ctx)),
			zap.Error(err),
		)
		return nil, err
	}

	err := s.Store.Del(pair.Key)
	if err != nil {
		if err == kvstore.ErrKeyDoesntExist {
			logger.Warn("DELETE failed: key does not exist",
				zap.String("key", pair.Key),
			)
			return nil, &pb.ErrKeyDoesntExist{}
		}
		logger.Error("Store DELETE failed",
			zap.String("key", pair.Key),
			zap.Error(err),
		)
		return nil, err
	}

	logger.Info("DELETE successful",
		zap.String("key", pair.Key),
	)
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
		logger.Error("Failed to register OpenCensus views", zap.Error(err))
		return nil, err
	}
	// END: metrics_traces

	// START: grpc_opts
	grpcOpts = append(grpcOpts,
		grpc.StreamInterceptor(
			grpc_middleware.ChainStreamServer(
				grpc_ctxtags.StreamServerInterceptor(),
				grpc_zap.StreamServerInterceptor(logger, zapOpts...),
				grpc_auth.StreamServerInterceptor(authenticate),
			)), grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_ctxtags.UnaryServerInterceptor(),
			grpc_zap.UnaryServerInterceptor(logger, zapOpts...),
			grpc_auth.UnaryServerInterceptor(authenticate),
		)),
		grpc.StatsHandler(&ocgrpc.ServerHandler{}),
	)

	gsrv := grpc.NewServer(grpcOpts...)
	srv, err := newgrpcServer(config)
	if err != nil {
		logger.Error("Failed to create gRPC server", zap.Error(err))
		return nil, err
	}

	hsrv := health.NewServer()
	hsrv.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(gsrv, hsrv)

	pb.RegisterKVStoreServiceServer(gsrv, srv)
	logger.Info("gRPC server initialized and ready")
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
