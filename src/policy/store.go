package policy

import (
	"sync"
)

// Policy represents an authorization policy
type Policy struct {
	Subject  string // Subject (user, etc.)
	Resource string // Resource
	Action   string // Action
	Allow    bool   // Whether to allow or deny
}

// Store represents a policy store
type Store struct {
	policies []Policy
	mu       sync.RWMutex
}

// NewStore creates a new policy store
func NewStore() *Store {
	return &Store{
		policies: make([]Policy, 0),
	}
}

// AddPolicy adds a policy to the store
func (s *Store) AddPolicy(subject, resource, action string, allow bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.policies = append(s.policies, Policy{
		Subject:  subject,
		Resource: resource,
		Action:   action,
		Allow:    allow,
	})
}

// CheckPolicy checks if a policy exists for the given subject, resource, and action
// If a policy exists, it returns the Allow value of that policy.
// If no policy exists, it returns false.
func (s *Store) CheckPolicy(subject, resource, action string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, p := range s.policies {
		if p.Subject == subject && p.Resource == resource && p.Action == action {
			return p.Allow
		}
	}

	return false
}

// ListPolicies returns all policies
func (s *Store) ListPolicies() []Policy {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Create a copy of the policies to return
	policies := make([]Policy, len(s.policies))
	copy(policies, s.policies)
	return policies
}

// FindSubjectsForResource finds subjects that are allowed to perform the given action on the given resource
func (s *Store) FindSubjectsForResource(resource, action string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	subjects := make([]string, 0)
	seen := make(map[string]bool)

	for _, p := range s.policies {
		if p.Resource == resource && p.Action == action && p.Allow {
			if !seen[p.Subject] {
				subjects = append(subjects, p.Subject)
				seen[p.Subject] = true
			}
		}
	}

	return subjects
}

// FindResourcesForSubject finds resources that the given subject is allowed to perform the given action on
func (s *Store) FindResourcesForSubject(subject, action string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	resources := make([]string, 0)
	seen := make(map[string]bool)

	for _, p := range s.policies {
		if p.Subject == subject && p.Action == action && p.Allow {
			if !seen[p.Resource] {
				resources = append(resources, p.Resource)
				seen[p.Resource] = true
			}
		}
	}

	return resources
}

// FindActionsForSubjectAndResource finds actions that the given subject is allowed to perform on the given resource
func (s *Store) FindActionsForSubjectAndResource(subject, resource string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	actions := make([]string, 0)
	seen := make(map[string]bool)

	for _, p := range s.policies {
		if p.Subject == subject && p.Resource == resource && p.Allow {
			if !seen[p.Action] {
				actions = append(actions, p.Action)
				seen[p.Action] = true
			}
		}
	}

	return actions
}
