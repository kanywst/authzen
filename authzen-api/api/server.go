package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/kanywst/authzen-api/policy"
)

// Server は Authorization API サーバーを表します
type Server struct {
	store *policy.Store
}

// AuthorizeRequest は認可リクエストの構造を定義します
type AuthorizeRequest struct {
	Principal struct {
		ID         string                 `json:"id"`
		Attributes map[string]interface{} `json:"attributes,omitempty"`
	} `json:"principal"`
	Resource struct {
		ID         string                 `json:"id"`
		Attributes map[string]interface{} `json:"attributes,omitempty"`
	} `json:"resource"`
	Action  string                 `json:"action"`
	Context map[string]interface{} `json:"context,omitempty"`
}

// AuthorizeResponse は認可レスポンスの構造を定義します
type AuthorizeResponse struct {
	Decision string `json:"decision"`
	Reason   string `json:"reason,omitempty"`
}

// EvaluationsRequest は複数の認可リクエストの構造を定義します
type EvaluationsRequest struct {
	Principal struct {
		ID         string                 `json:"id"`
		Attributes map[string]interface{} `json:"attributes,omitempty"`
	} `json:"principal"`
	Context     map[string]interface{} `json:"context,omitempty"`
	Action      string                 `json:"action,omitempty"`
	Evaluations []struct {
		Resource struct {
			ID         string                 `json:"id"`
			Attributes map[string]interface{} `json:"attributes,omitempty"`
		} `json:"resource"`
		Action string `json:"action,omitempty"`
	} `json:"evaluations"`
	Semantic string `json:"semantic,omitempty"` // permit_overrides, deny_overrides, permit_on_first_permit, deny_on_first_deny
}

// EvaluationsResponse は複数の認可レスポンスの構造を定義します
type EvaluationsResponse struct {
	Evaluations []struct {
		Decision string                 `json:"decision"`
		Reason   string                 `json:"reason,omitempty"`
		Context  map[string]interface{} `json:"context,omitempty"`
	} `json:"evaluations"`
}

// NewServer は新しい API サーバーを作成します
func NewServer(store *policy.Store) *Server {
	return &Server{
		store: store,
	}
}

// Router は API ルーターを設定して返します
func (s *Server) Router() http.Handler {
	r := mux.NewRouter()

	// 認可エンドポイント
	r.HandleFunc("/v1/authorize", s.handleAuthorize).Methods("POST")

	// 複数認可エンドポイント
	r.HandleFunc("/v1/evaluations", s.handleEvaluations).Methods("POST")

	// 検索エンドポイント
	r.HandleFunc("/v1/search/subject", s.handleSearchSubject).Methods("POST")
	r.HandleFunc("/v1/search/resource", s.handleSearchResource).Methods("POST")
	r.HandleFunc("/v1/search/action", s.handleSearchAction).Methods("POST")

	// ポリシー一覧エンドポイント
	r.HandleFunc("/v1/policies", s.handleListPolicies).Methods("GET")

	// ヘルスチェックエンドポイント
	r.HandleFunc("/health", s.handleHealth).Methods("GET")

	return r
}

// handleAuthorize は認可リクエストを処理します
func (s *Server) handleAuthorize(w http.ResponseWriter, r *http.Request) {
	var req AuthorizeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// ポリシーの評価
	allowed := s.store.CheckPolicy(req.Principal.ID, req.Resource.ID, req.Action)

	// レスポンスの作成
	resp := AuthorizeResponse{}
	if allowed {
		resp.Decision = "ALLOW"
	} else {
		resp.Decision = "DENY"
		resp.Reason = "Access denied by policy"
	}

	// レスポンスの送信
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// handleListPolicies はポリシー一覧を返します
func (s *Server) handleListPolicies(w http.ResponseWriter, r *http.Request) {
	policies := s.store.ListPolicies()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(policies)
}

// handleHealth はヘルスチェックリクエストを処理します
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// handleEvaluations は複数の認可リクエストを処理します
func (s *Server) handleEvaluations(w http.ResponseWriter, r *http.Request) {
	var req EvaluationsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// レスポンスの準備
	resp := EvaluationsResponse{
		Evaluations: make([]struct {
			Decision string                 `json:"decision"`
			Reason   string                 `json:"reason,omitempty"`
			Context  map[string]interface{} `json:"context,omitempty"`
		}, len(req.Evaluations)),
	}

	// 各評価リクエストを処理
	for i, eval := range req.Evaluations {
		// アクションがリクエストレベルで指定されている場合は、それを使用
		action := eval.Action
		if action == "" {
			action = req.Action
		}

		// ポリシーの評価
		allowed := s.store.CheckPolicy(req.Principal.ID, eval.Resource.ID, action)

		// 結果の設定
		if allowed {
			resp.Evaluations[i].Decision = "ALLOW"
		} else {
			resp.Evaluations[i].Decision = "DENY"
			resp.Evaluations[i].Reason = "Access denied by policy"
		}
	}

	// セマンティクスに基づく処理（オプション）
	if req.Semantic != "" {
		// セマンティクスに基づいて結果を調整する処理を実装可能
		// 例: permit_overrides, deny_overrides, permit_on_first_permit, deny_on_first_deny
	}

	// レスポンスの送信
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// SubjectSearchRequest はSubject検索リクエストの構造を定義します
type SubjectSearchRequest struct {
	Resource struct {
		ID         string                 `json:"id"`
		Attributes map[string]interface{} `json:"attributes,omitempty"`
	} `json:"resource,omitempty"`
	Action  string                 `json:"action,omitempty"`
	Context map[string]interface{} `json:"context,omitempty"`
}

// SubjectSearchResponse はSubject検索レスポンスの構造を定義します
type SubjectSearchResponse struct {
	Results []struct {
		ID         string                 `json:"id"`
		Attributes map[string]interface{} `json:"attributes,omitempty"`
	} `json:"results"`
	Page struct {
		NextToken string `json:"next_token,omitempty"`
	} `json:"page,omitempty"`
}

// handleSearchSubject はSubject検索リクエストを処理します
func (s *Server) handleSearchSubject(w http.ResponseWriter, r *http.Request) {
	var req SubjectSearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 実際の実装では、ポリシーストアから条件に合致するSubjectを検索
	// このサンプルでは、ダミーデータを返す
	resp := SubjectSearchResponse{
		Results: []struct {
			ID         string                 `json:"id"`
			Attributes map[string]interface{} `json:"attributes,omitempty"`
		}{
			{
				ID: "user:alice",
				Attributes: map[string]interface{}{
					"name": "Alice",
				},
			},
			{
				ID: "user:bob",
				Attributes: map[string]interface{}{
					"name": "Bob",
				},
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// ResourceSearchRequest はResource検索リクエストの構造を定義します
type ResourceSearchRequest struct {
	Principal struct {
		ID         string                 `json:"id"`
		Attributes map[string]interface{} `json:"attributes,omitempty"`
	} `json:"principal"`
	Action  string                 `json:"action,omitempty"`
	Context map[string]interface{} `json:"context,omitempty"`
}

// ResourceSearchResponse はResource検索レスポンスの構造を定義します
type ResourceSearchResponse struct {
	Results []struct {
		ID         string                 `json:"id"`
		Attributes map[string]interface{} `json:"attributes,omitempty"`
	} `json:"results"`
	Page struct {
		NextToken string `json:"next_token,omitempty"`
	} `json:"page,omitempty"`
}

// handleSearchResource はResource検索リクエストを処理します
func (s *Server) handleSearchResource(w http.ResponseWriter, r *http.Request) {
	var req ResourceSearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 実際の実装では、ポリシーストアから条件に合致するResourceを検索
	// このサンプルでは、ダミーデータを返す
	resp := ResourceSearchResponse{
		Results: []struct {
			ID         string                 `json:"id"`
			Attributes map[string]interface{} `json:"attributes,omitempty"`
		}{
			{
				ID: "document:report",
				Attributes: map[string]interface{}{
					"type": "document",
				},
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// ActionSearchRequest はAction検索リクエストの構造を定義します
type ActionSearchRequest struct {
	Principal struct {
		ID         string                 `json:"id"`
		Attributes map[string]interface{} `json:"attributes,omitempty"`
	} `json:"principal,omitempty"`
	Resource struct {
		ID         string                 `json:"id"`
		Attributes map[string]interface{} `json:"attributes,omitempty"`
	} `json:"resource,omitempty"`
	Context map[string]interface{} `json:"context,omitempty"`
}

// ActionSearchResponse はAction検索レスポンスの構造を定義します
type ActionSearchResponse struct {
	Results []string `json:"results"`
	Page    struct {
		NextToken string `json:"next_token,omitempty"`
	} `json:"page,omitempty"`
}

// handleSearchAction はAction検索リクエストを処理します
func (s *Server) handleSearchAction(w http.ResponseWriter, r *http.Request) {
	var req ActionSearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 実際の実装では、ポリシーストアから条件に合致するActionを検索
	// このサンプルでは、ダミーデータを返す
	resp := ActionSearchResponse{
		Results: []string{"read", "write"},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
