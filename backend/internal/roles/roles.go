package roles

// User maps to the persisted end-user role name kept for compatibility.
const (
	User       = "customer"
	Admin      = "admin"
	SuperAdmin = "superadmin"
)

func UserPortal() []string {
	return []string{User}
}

func AdminPortal() []string {
	return []string{Admin, SuperAdmin}
}
