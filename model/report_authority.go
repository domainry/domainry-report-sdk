package reportmodel

import (
	"fmt"
	"strings"

	identitysdk "github.com/domainry/domainry-identity-sdk"
)

// ReportAuthority is the caller proof accepted by Report business use cases.
// Browser/API calls carry AccessToken. Trusted embedded orchestration may
// carry a host-resolved Subject; remote bindings must ignore Subject and send
// only the original access token to the Report service.
type ReportAuthority struct {
	AccessToken        string         `json:"-"`
	RequestID          string         `json:"-"`
	BusinessProfileKey string         `json:"-"`
	BusinessProfileID  string         `json:"-"`
	Subject            *ReportSubject `json:"-"`
}

func (a ReportAuthority) Validate() error {
	if strings.TrimSpace(a.AccessToken) != "" {
		return nil
	}
	if a.Subject != nil {
		return a.Subject.Validate()
	}
	return fmt.Errorf("Report authority is required")
}

type ReportBusinessProfile struct {
	BindingKey string                         `json:"binding_key"`
	ObjectKey  string                         `json:"object_key"`
	ProfileID  string                         `json:"profile_id"`
	Claims     map[string]ReportBusinessClaim `json:"claims,omitempty"`
}

type ReportBusinessClaim struct {
	Type  string `json:"type"`
	Value any    `json:"value"`
}

// ReportSubject contains only the authenticated facts Report needs for its
// own authorization and consistency rules. The host computes
// AccessScopeHash from every RLS/CLS-affecting fact; Report never attempts to
// reconstruct host business-profile policy from request headers.
type ReportSubject struct {
	Principal             identitysdk.Principal          `json:"principal"`
	RequestID             string                         `json:"request_id,omitempty"`
	CorrelationID         string                         `json:"correlation_id,omitempty"`
	CausationID           string                         `json:"causation_id,omitempty"`
	BusinessProfiles      []ReportBusinessProfile        `json:"business_profiles,omitempty"`
	ActiveBusinessProfile *ReportBusinessProfile         `json:"active_business_profile,omitempty"`
	BusinessClaims        map[string]ReportBusinessClaim `json:"business_claims,omitempty"`
	AccessScopeHash       string                         `json:"access_scope_hash"`
	// TrustedProcess marks an in-process authority resolved by the embedding
	// host. It is deliberately excluded from JSON so remote callers cannot
	// manufacture process authority. ProcessCapabilities must contain the exact
	// actions and data permissions derived from the work item being executed.
	TrustedProcess      bool     `json:"-"`
	ProcessCapabilities []string `json:"-"`
}

func (s ReportSubject) Validate() error {
	if !s.Principal.Known || strings.TrimSpace(s.Principal.WorkspaceID) == "" || strings.TrimSpace(s.Principal.UserID) == "" {
		return fmt.Errorf("Report subject is not authenticated")
	}
	if strings.TrimSpace(s.AccessScopeHash) == "" {
		return fmt.Errorf("Report subject access scope hash is required")
	}
	return nil
}

func (s ReportSubject) HasPermission(permission string) bool {
	permission = strings.TrimSpace(permission)
	if s.Principal.HasPermission(permission) {
		return true
	}
	if !s.TrustedProcess || permission == "" {
		return false
	}
	for _, capability := range s.ProcessCapabilities {
		if strings.TrimSpace(capability) == permission {
			return true
		}
	}
	return false
}

func (s ReportSubject) HasAllPermissions(permissions []string) bool {
	if len(permissions) == 0 {
		return false
	}
	for _, permission := range permissions {
		if !s.HasPermission(permission) {
			return false
		}
	}
	return true
}

// WithExactProcessCapabilities adds owner-derived capabilities only to a
// trusted in-process subject. Wildcard-shaped entries are ignored and an
// externally resolved subject is returned unchanged.
func (s ReportSubject) WithExactProcessCapabilities(capabilities ...string) ReportSubject {
	if !s.TrustedProcess {
		return s
	}
	result := s
	result.ProcessCapabilities = append([]string(nil), s.ProcessCapabilities...)
	seen := make(map[string]bool, len(result.ProcessCapabilities)+len(capabilities))
	for _, capability := range result.ProcessCapabilities {
		seen[strings.TrimSpace(capability)] = true
	}
	for _, capability := range capabilities {
		capability = strings.TrimSpace(capability)
		if capability == "" || strings.Contains(capability, "*") || seen[capability] {
			continue
		}
		seen[capability] = true
		result.ProcessCapabilities = append(result.ProcessCapabilities, capability)
	}
	return result
}
