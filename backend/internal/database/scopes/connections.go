package scopes

import "gorm.io/gorm"

const (
	// PortalScopeKey is attached to scoped gorm sessions for observability and future per-scope db routing.
	PortalScopeKey = "portal_scope"
	ScopePublic    = "public"
	ScopeUser      = "user_portal"
	ScopeAdmin     = "admin_portal"
	ScopeAuth      = "auth_scope"
	ScopeSystem    = "system"
)

// Connections holds one DB session per application scope.
type Connections struct {
	Public      *gorm.DB
	UserPortal  *gorm.DB
	AdminPortal *gorm.DB
	AuthScope   *gorm.DB
	System      *gorm.DB
}

// NewConnections creates isolated sessions per scope.
func NewConnections(base *gorm.DB) *Connections {
	if base == nil {
		return &Connections{}
	}

	return &Connections{
		Public:      scopedSession(base, ScopePublic),
		UserPortal:  scopedSession(base, ScopeUser),
		AdminPortal: scopedSession(base, ScopeAdmin),
		AuthScope:   scopedSession(base, ScopeAuth),
		System:      scopedSession(base, ScopeSystem),
	}
}

func scopedSession(base *gorm.DB, scope string) *gorm.DB {
	return base.Session(&gorm.Session{NewDB: true}).Set(PortalScopeKey, scope)
}
