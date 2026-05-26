package order

import (
	modelsOrder "candypro/api/internal/models/order"
	"candypro/api/internal/pkg/safego"
	"context"
	"log"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"
)

// defaultPolicyRefreshInterval bounds how stale the in-process cache can become
// when another instance edits the policies table directly. 5 minutes balances
// admin-edit latency against DB load (M-20).
const defaultPolicyRefreshInterval = 5 * time.Minute

// CountryPaymentPolicyService provides table-driven country payment policy lookups.
type CountryPaymentPolicyService struct {
	db          *gorm.DB
	cache       map[string]*modelsOrder.CountryPaymentPolicy
	mu          sync.RWMutex
	refreshOnce sync.Once
}

// NewCountryPaymentPolicyService creates a new CountryPaymentPolicyService.
func NewCountryPaymentPolicyService(db *gorm.DB) *CountryPaymentPolicyService {
	svc := &CountryPaymentPolicyService{db: db}
	svc.loadPolicies()
	svc.startAutoRefresh(defaultPolicyRefreshInterval)
	return svc
}

// startAutoRefresh launches a single background goroutine that periodically
// reloads the policy cache so multi-instance deployments converge after direct
// DB edits or hot config rollouts (M-20). Idempotent across calls.
func (s *CountryPaymentPolicyService) startAutoRefresh(interval time.Duration) {
	if interval <= 0 {
		return
	}
	s.refreshOnce.Do(func() {
		safego.Go("country-payment-policy.refresh", func() {
			ticker := time.NewTicker(interval)
			defer ticker.Stop()
			for range ticker.C {
				s.loadPolicies()
			}
		})
	})
}

// loadPolicies loads all policies from the database into memory.
func (s *CountryPaymentPolicyService) loadPolicies() {
	var policies []modelsOrder.CountryPaymentPolicy
	if err := s.db.Find(&policies).Error; err != nil {
		log.Printf("Warning: failed to load country payment policies: %v", err)
		return
	}
	s.mu.Lock()
	s.cache = make(map[string]*modelsOrder.CountryPaymentPolicy, len(policies))
	for i := range policies {
		s.cache[policies[i].Country] = &policies[i]
	}
	s.mu.Unlock()
}

// Reload refreshes the in-memory cache from the database.
func (s *CountryPaymentPolicyService) Reload() {
	s.loadPolicies()
}

// RequiresFullPrepayment returns true if the country requires full prepayment.
// Falls back to false for unknown countries.
func (s *CountryPaymentPolicyService) RequiresFullPrepayment(country string) bool {
	canonical := normalizeCountryPolicyKey(country)
	s.mu.RLock()
	policy, ok := s.cache[canonical]
	s.mu.RUnlock()
	if !ok || policy == nil {
		return false
	}
	return policy.RequiresFullPrepayment
}

// GetPolicy returns the payment policy for a country, or nil if none exists.
func (s *CountryPaymentPolicyService) GetPolicy(country string) *modelsOrder.CountryPaymentPolicy {
	canonical := normalizeCountryPolicyKey(country)
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cache[canonical]
}

// ListPolicies returns all configured payment policies.
func (s *CountryPaymentPolicyService) ListPolicies(ctx context.Context) ([]modelsOrder.CountryPaymentPolicy, error) {
	var policies []modelsOrder.CountryPaymentPolicy
	if err := s.db.WithContext(ctx).Order("country ASC").Find(&policies).Error; err != nil {
		return nil, err
	}
	return policies, nil
}

// UpsertPolicy creates or updates a country payment policy.
func (s *CountryPaymentPolicyService) UpsertPolicy(ctx context.Context, policy *modelsOrder.CountryPaymentPolicy) error {
	canonical := normalizeCountryPolicyKey(policy.Country)
	policy.Country = canonical
	if err := s.db.WithContext(ctx).Where("country = ?", canonical).FirstOrCreate(policy).Error; err != nil {
		return err
	}
	s.loadPolicies()
	return nil
}

// DeletePolicy removes a country payment policy.
func (s *CountryPaymentPolicyService) DeletePolicy(ctx context.Context, country string) error {
	canonical := normalizeCountryPolicyKey(country)
	if err := s.db.WithContext(ctx).Where("country = ?", canonical).Delete(&modelsOrder.CountryPaymentPolicy{}).Error; err != nil {
		return err
	}
	s.loadPolicies()
	return nil
}

func normalizeCountryPolicyKey(country string) string {
	normalized := strings.ToLower(strings.TrimSpace(country))
	switch normalized {
	case "in", "india", "bharat":
		return "india"
	case "pk", "pakistan", "islamic republic of pakistan":
		return "pakistan"
	default:
		return normalized
	}
}
