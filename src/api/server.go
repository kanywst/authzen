package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"authzen/policy"

	"github.com/gorilla/mux"
)

// Server represents an Authorization API server
type Server struct {
	store    *policy.Store
	baseURL  string
	router   *mux.Router
	handlers map[string]http.HandlerFunc
}

// NewServer creates a new API server
func NewServer(store *policy.Store, baseURL string) *Server {
	s := &Server{
		store:    store,
		baseURL:  baseURL,
		handlers: make(map[string]http.HandlerFunc),
	}

	// Initialize router
	s.router = mux.NewRouter()

	// Register handlers
	s.registerHandlers()

	return s
}

// Router returns the API router
func (s *Server) Router() http.Handler {
	return s.router
}

// registerHandlers registers API handlers
func (s *Server) registerHandlers() {
	// Metadata discovery endpoint
	s.router.HandleFunc("/.well-known/authzen-configuration", s.handleMetadata).Methods("GET")

	// Authorization endpoints
	s.router.HandleFunc("/access/v1/evaluation", s.handleAuthorize).Methods("POST")
	s.router.HandleFunc("/access/v1/evaluations", s.handleEvaluations).Methods("POST")

	// Search endpoints
	s.router.HandleFunc("/access/v1/search/subject", s.handleSearchSubject).Methods("POST")
	s.router.HandleFunc("/access/v1/search/resource", s.handleSearchResource).Methods("POST")
	s.router.HandleFunc("/access/v1/search/action", s.handleSearchAction).Methods("POST")

	// Policy listing endpoint
	s.router.HandleFunc("/v1/policies", s.handleListPolicies).Methods("GET")

	// Health check endpoint
	s.router.HandleFunc("/health", s.handleHealth).Methods("GET")
}

// handleMetadata handles metadata discovery requests
func (s *Server) handleMetadata(w http.ResponseWriter, r *http.Request) {
	metadata := MetadataResponse{
		PolicyDecisionPoint:       s.baseURL,
		AccessEvaluationEndpoint:  fmt.Sprintf("%s/access/v1/evaluation", s.baseURL),
		AccessEvaluationsEndpoint: fmt.Sprintf("%s/access/v1/evaluations", s.baseURL),
		SearchSubjectEndpoint:     fmt.Sprintf("%s/access/v1/search/subject", s.baseURL),
		SearchResourceEndpoint:    fmt.Sprintf("%s/access/v1/search/resource", s.baseURL),
		SearchActionEndpoint:      fmt.Sprintf("%s/access/v1/search/action", s.baseURL),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metadata)
}

// handleAuthorize handles authorization requests
func (s *Server) handleAuthorize(w http.ResponseWriter, r *http.Request) {
	var req AuthorizeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if err := validateAuthorizeRequest(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Evaluate policy
	allowed := s.store.CheckPolicy(req.Subject.ID, req.Resource.ID, req.Action.Name)

	// Create response
	resp := AuthorizeResponse{}
	if allowed {
		resp.Decision = "ALLOW"
	} else {
		resp.Decision = "DENY"
		resp.Context = map[string]interface{}{
			"reason": "Access denied by policy",
		}
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// handleEvaluations handles multiple authorization requests
func (s *Server) handleEvaluations(w http.ResponseWriter, r *http.Request) {
	var req EvaluationsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if err := validateEvaluationsRequest(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Prepare response
	resp := EvaluationsResponse{
		Evaluations: make([]EvaluationResult, len(req.Evaluations)),
	}

	// Get evaluation semantics
	semantic := req.Options.EvaluationsSemantic
	if semantic == "" {
		semantic = "execute_all" // Default
	}

	// Process each evaluation request
	for i, eval := range req.Evaluations {
		// Use action from request level if specified
		action := eval.Action.Name
		if action == "" && req.Action.Name != "" {
			action = req.Action.Name
		}

		// Evaluate policy
		allowed := s.store.CheckPolicy(req.Subject.ID, eval.Resource.ID, action)

		// Set result
		if allowed {
			resp.Evaluations[i].Decision = "ALLOW"
		} else {
			resp.Evaluations[i].Decision = "DENY"
			resp.Evaluations[i].Context = map[string]interface{}{
				"reason": "Access denied by policy",
			}
		}

		// Process based on semantics
		if semantic == "deny_on_first_deny" && !allowed {
			// Stop on first deny
			resp.Evaluations = resp.Evaluations[:i+1]
			break
		} else if semantic == "permit_on_first_permit" && allowed {
			// Stop on first permit
			resp.Evaluations = resp.Evaluations[:i+1]
			break
		}
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// handleSearchSubject handles Subject search requests
func (s *Server) handleSearchSubject(w http.ResponseWriter, r *http.Request) {
	var req SubjectSearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if err := validateSubjectSearchRequest(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Search for subjects
	subjects := s.store.FindSubjectsForResource(req.Resource.ID, req.Action.Name)

	// Create response
	resp := SubjectSearchResponse{
		Results: make([]Subject, 0),
		Page: struct {
			NextToken string `json:"next_token,omitempty"`
		}{
			NextToken: "",
		},
	}

	// Add results
	for _, subjectID := range subjects {
		parts := strings.SplitN(subjectID, ":", 2)
		if len(parts) != 2 {
			continue
		}

		resp.Results = append(resp.Results, Subject{
			Type: parts[0],
			ID:   subjectID,
		})
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// handleSearchResource handles Resource search requests
func (s *Server) handleSearchResource(w http.ResponseWriter, r *http.Request) {
	var req ResourceSearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if err := validateResourceSearchRequest(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Search for resources
	resources := s.store.FindResourcesForSubject(req.Subject.ID, req.Action.Name)

	// Create response
	resp := ResourceSearchResponse{
		Results: make([]Resource, 0),
		Page: struct {
			NextToken string `json:"next_token,omitempty"`
		}{
			NextToken: "",
		},
	}

	// Add results
	for _, resourceID := range resources {
		parts := strings.SplitN(resourceID, ":", 2)
		if len(parts) != 2 {
			continue
		}

		resp.Results = append(resp.Results, Resource{
			Type: parts[0],
			ID:   resourceID,
		})
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// handleSearchAction handles Action search requests
func (s *Server) handleSearchAction(w http.ResponseWriter, r *http.Request) {
	var req ActionSearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if err := validateActionSearchRequest(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Search for actions
	actions := s.store.FindActionsForSubjectAndResource(req.Subject.ID, req.Resource.ID)

	// Create response
	resp := ActionSearchResponse{
		Results: make([]Action, 0),
		Page: struct {
			NextToken string `json:"next_token,omitempty"`
		}{
			NextToken: "",
		},
	}

	// Add results
	for _, actionName := range actions {
		resp.Results = append(resp.Results, Action{
			Name: actionName,
		})
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// handleListPolicies returns a list of policies
func (s *Server) handleListPolicies(w http.ResponseWriter, r *http.Request) {
	policies := s.store.ListPolicies()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(policies)
}

// handleHealth handles health check requests
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// validateAuthorizeRequest validates an authorization request
func validateAuthorizeRequest(req AuthorizeRequest) error {
	if req.Subject.Type == "" || req.Subject.ID == "" {
		return fmt.Errorf("subject type and id are required")
	}
	if req.Resource.Type == "" || req.Resource.ID == "" {
		return fmt.Errorf("resource type and id are required")
	}
	if req.Action.Name == "" {
		return fmt.Errorf("action name is required")
	}
	return nil
}

// validateEvaluationsRequest validates multiple authorization requests
func validateEvaluationsRequest(req EvaluationsRequest) error {
	if req.Subject.Type == "" || req.Subject.ID == "" {
		return fmt.Errorf("subject type and id are required")
	}
	if len(req.Evaluations) == 0 {
		return fmt.Errorf("at least one evaluation is required")
	}
	for i, eval := range req.Evaluations {
		if eval.Resource.Type == "" || eval.Resource.ID == "" {
			return fmt.Errorf("resource type and id are required for evaluation %d", i)
		}
		// Action is required for each evaluation if not specified at the request level
		if req.Action.Name == "" && eval.Action.Name == "" {
			return fmt.Errorf("action name is required for evaluation %d", i)
		}
	}
	// Validate evaluation semantics
	if req.Options.EvaluationsSemantic != "" {
		semantic := req.Options.EvaluationsSemantic
		if semantic != "execute_all" && semantic != "deny_on_first_deny" && semantic != "permit_on_first_permit" {
			return fmt.Errorf("invalid evaluations_semantic: %s", semantic)
		}
	}
	return nil
}

// validateSubjectSearchRequest validates a Subject search request
func validateSubjectSearchRequest(req SubjectSearchRequest) error {
	if req.Subject.Type == "" {
		return fmt.Errorf("subject type is required")
	}
	if req.Resource.Type == "" || req.Resource.ID == "" {
		return fmt.Errorf("resource type and id are required")
	}
	if req.Action.Name == "" {
		return fmt.Errorf("action name is required")
	}
	return nil
}

// validateResourceSearchRequest validates a Resource search request
func validateResourceSearchRequest(req ResourceSearchRequest) error {
	if req.Subject.Type == "" || req.Subject.ID == "" {
		return fmt.Errorf("subject type and id are required")
	}
	if req.Resource.Type == "" {
		return fmt.Errorf("resource type is required")
	}
	if req.Action.Name == "" {
		return fmt.Errorf("action name is required")
	}
	return nil
}

// validateActionSearchRequest validates an Action search request
func validateActionSearchRequest(req ActionSearchRequest) error {
	if req.Subject.Type == "" || req.Subject.ID == "" {
		return fmt.Errorf("subject type and id are required")
	}
	if req.Resource.Type == "" || req.Resource.ID == "" {
		return fmt.Errorf("resource type and id are required")
	}
	return nil
}
