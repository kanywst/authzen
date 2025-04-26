package policy

import (
	"fmt"
	"sync"
)

// Policy は認可ポリシーを表します
type Policy struct {
	Principal string `json:"principal"`
	Resource  string `json:"resource"`
	Action    string `json:"action"`
	Allow     bool   `json:"allow"`
}

// Store はポリシーストアを表します
type Store struct {
	policies []Policy
	mu       sync.RWMutex
}

// NewStore は新しいポリシーストアを作成します
func NewStore() *Store {
	return &Store{
		policies: make([]Policy, 0),
	}
}

// AddPolicy はポリシーを追加します
func (s *Store) AddPolicy(principal, resource, action string, allow bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.policies = append(s.policies, Policy{
		Principal: principal,
		Resource:  resource,
		Action:    action,
		Allow:     allow,
	})
}

// CheckPolicy はポリシーを評価します
func (s *Store) CheckPolicy(principal, resource, action string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, p := range s.policies {
		if p.Principal == principal && p.Resource == resource && p.Action == action {
			return p.Allow
		}
	}

	// デフォルトでは拒否
	return false
}

// ListPolicies は全てのポリシーを返します
func (s *Store) ListPolicies() []Policy {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// ポリシーのコピーを作成して返す
	policies := make([]Policy, len(s.policies))
	copy(policies, s.policies)
	return policies
}

// String はポリシーの文字列表現を返します
func (p Policy) String() string {
	decision := "DENY"
	if p.Allow {
		decision = "ALLOW"
	}
	return fmt.Sprintf("%s -> %s -> %s: %s", p.Principal, p.Resource, p.Action, decision)
}
