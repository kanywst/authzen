package api

// Subject represents a principal (user or machine principal)
type Subject struct {
	Type       string                 `json:"type"`
	ID         string                 `json:"id"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// Resource represents a resource
type Resource struct {
	Type       string                 `json:"type"`
	ID         string                 `json:"id"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// Action represents an action
type Action struct {
	Name       string                 `json:"name"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// Context represents context information
type Context map[string]interface{}

// AuthorizeRequest represents an authorization request
type AuthorizeRequest struct {
	Subject  Subject  `json:"subject"`
	Resource Resource `json:"resource"`
	Action   Action   `json:"action"`
	Context  Context  `json:"context,omitempty"`
}

// AuthorizeResponse represents an authorization response
type AuthorizeResponse struct {
	Decision string                 `json:"decision"`
	Context  map[string]interface{} `json:"context,omitempty"`
}

// EvaluationItem represents an evaluation item
type EvaluationItem struct {
	Resource Resource `json:"resource"`
	Action   Action   `json:"action,omitempty"`
}

// EvaluationsRequest represents multiple authorization requests
type EvaluationsRequest struct {
	Subject     Subject          `json:"subject"`
	Action      Action           `json:"action,omitempty"`
	Context     Context          `json:"context,omitempty"`
	Evaluations []EvaluationItem `json:"evaluations"`
	Options     struct {
		EvaluationsSemantic string `json:"evaluations_semantic,omitempty"`
	} `json:"options,omitempty"`
}

// EvaluationResult represents an evaluation result
type EvaluationResult struct {
	Decision string                 `json:"decision"`
	Context  map[string]interface{} `json:"context,omitempty"`
}

// EvaluationsResponse represents multiple authorization responses
type EvaluationsResponse struct {
	Evaluations []EvaluationResult `json:"evaluations"`
}

// SubjectSearchRequest represents a Subject search request
type SubjectSearchRequest struct {
	Subject  Subject  `json:"subject"`
	Resource Resource `json:"resource"`
	Action   Action   `json:"action"`
	Context  Context  `json:"context,omitempty"`
	Page     struct {
		NextToken string `json:"next_token,omitempty"`
	} `json:"page,omitempty"`
}

// SubjectSearchResponse represents a Subject search response
type SubjectSearchResponse struct {
	Results []Subject `json:"results"`
	Page    struct {
		NextToken string `json:"next_token,omitempty"`
	} `json:"page,omitempty"`
}

// ResourceSearchRequest represents a Resource search request
type ResourceSearchRequest struct {
	Subject  Subject  `json:"subject"`
	Resource Resource `json:"resource"`
	Action   Action   `json:"action"`
	Context  Context  `json:"context,omitempty"`
	Page     struct {
		NextToken string `json:"next_token,omitempty"`
	} `json:"page,omitempty"`
}

// ResourceSearchResponse represents a Resource search response
type ResourceSearchResponse struct {
	Results []Resource `json:"results"`
	Page    struct {
		NextToken string `json:"next_token,omitempty"`
	} `json:"page,omitempty"`
}

// ActionSearchRequest represents an Action search request
type ActionSearchRequest struct {
	Subject  Subject  `json:"subject"`
	Resource Resource `json:"resource"`
	Context  Context  `json:"context,omitempty"`
	Page     struct {
		NextToken string `json:"next_token,omitempty"`
	} `json:"page,omitempty"`
}

// ActionSearchResponse represents an Action search response
type ActionSearchResponse struct {
	Results []Action `json:"results"`
	Page    struct {
		NextToken string `json:"next_token,omitempty"`
	} `json:"page,omitempty"`
}

// MetadataResponse represents a metadata response
type MetadataResponse struct {
	PolicyDecisionPoint       string `json:"policy_decision_point"`
	AccessEvaluationEndpoint  string `json:"access_evaluation_endpoint"`
	AccessEvaluationsEndpoint string `json:"access_evaluations_endpoint,omitempty"`
	SearchSubjectEndpoint     string `json:"search_subject_endpoint,omitempty"`
	SearchResourceEndpoint    string `json:"search_resource_endpoint,omitempty"`
	SearchActionEndpoint      string `json:"search_action_endpoint,omitempty"`
}
