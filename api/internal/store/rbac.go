package store

import "fmt"

// Team roles (ascending privilege).
const (
	RoleViewer = "viewer"
	RoleMember = "member"
	RoleAdmin  = "admin"
	RoleOwner  = "owner"
)

var roleRank = map[string]int{
	RoleViewer: 1,
	RoleMember: 2,
	RoleAdmin:  3,
	RoleOwner:  4,
}

// RoleAtLeast reports whether have is at least as privileged as need.
func RoleAtLeast(have, need string) bool {
	return roleRank[have] >= roleRank[need] && roleRank[need] > 0
}

// ValidRole reports whether role is a known team role.
func ValidRole(role string) bool {
	_, ok := roleRank[role]
	return ok
}

// TeamMember is a user's membership on a team.
type TeamMember struct {
	UserID string `json:"userId"`
	Email  string `json:"email"`
	Role   string `json:"role"`
}

// ErrNotMember is returned when a membership lookup fails.
var ErrNotMember = fmt.Errorf("not a team member")

// ErrLastOwner prevents removing or demoting the last owner.
var ErrLastOwner = fmt.Errorf("cannot remove or demote the last owner")
