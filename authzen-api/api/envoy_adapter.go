package api

import (
	"context"
	"log"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	auth "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	envoy_type "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"github.com/kanywst/authzen-api/policy"
	"google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

// EnvoyAuthServer implements the Envoy external authorization gRPC service
type EnvoyAuthServer struct {
	auth.UnimplementedAuthorizationServer
	store *policy.Store
}

// NewEnvoyAuthServer creates a new EnvoyAuthServer
func NewEnvoyAuthServer(store *policy.Store) *EnvoyAuthServer {
	return &EnvoyAuthServer{
		store: store,
	}
}

// Register registers the EnvoyAuthServer with a gRPC server
func (s *EnvoyAuthServer) Register(grpcServer *grpc.Server) {
	auth.RegisterAuthorizationServer(grpcServer, s)
}

// Check implements the Check method of the authorization service
func (s *EnvoyAuthServer) Check(ctx context.Context, req *auth.CheckRequest) (*auth.CheckResponse, error) {
	log.Printf("Received authorization check request: %v", req)

	// リクエストからprincipal、resource、actionを抽出
	attrs := req.GetAttributes()
	if attrs == nil {
		return denied("No attributes found in the request", codes.InvalidArgument), nil
	}

	// HTTPリクエストの情報を取得
	http := attrs.GetRequest().GetHttp()
	if http == nil {
		return denied("No HTTP request found", codes.InvalidArgument), nil
	}

	// ヘッダーからprincipalを取得
	principalID := http.GetHeaders()["x-user-id"]
	if principalID == "" {
		return denied("No user ID found in headers", codes.Unauthenticated), nil
	}

	// パスからresourceを取得
	resourceID := "resource:" + http.GetPath()

	// メソッドからactionを取得
	action := http.GetMethod()

	// 認可チェック
	allow := s.store.CheckPolicy(principalID, resourceID, action)
	if !allow {
		return denied("Access denied by policy", codes.PermissionDenied), nil
	}

	// アクセス許可
	return &auth.CheckResponse{
		Status: &status.Status{
			Code: int32(codes.OK),
		},
		HttpResponse: &auth.CheckResponse_OkResponse{
			OkResponse: &auth.OkHttpResponse{
				Headers: []*core.HeaderValueOption{
					{
						Header: &core.HeaderValue{
							Key:   "x-authzen-result",
							Value: "allowed",
						},
					},
				},
			},
		},
	}, nil
}

// denied creates a denied check response
func denied(message string, code codes.Code) *auth.CheckResponse {
	return &auth.CheckResponse{
		Status: &status.Status{
			Code:    int32(code),
			Message: message,
		},
		HttpResponse: &auth.CheckResponse_DeniedResponse{
			DeniedResponse: &auth.DeniedHttpResponse{
				Status: &envoy_type.HttpStatus{
					Code: envoy_type.StatusCode_Forbidden,
				},
				Body: message,
			},
		},
	}
}
