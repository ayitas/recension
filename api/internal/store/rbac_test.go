package store_test

import (
	"testing"

	"github.com/ayitas/recension/api/internal/store"
)

func TestRoleAtLeast(t *testing.T) {
	cases := []struct {
		have, need string
		ok         bool
	}{
		{store.RoleOwner, store.RoleAdmin, true},
		{store.RoleAdmin, store.RoleMember, true},
		{store.RoleMember, store.RoleViewer, true},
		{store.RoleViewer, store.RoleMember, false},
		{store.RoleMember, store.RoleAdmin, false},
		{"", store.RoleViewer, false},
		{store.RoleViewer, "nope", false},
	}
	for _, tc := range cases {
		if got := store.RoleAtLeast(tc.have, tc.need); got != tc.ok {
			t.Fatalf("RoleAtLeast(%q,%q)=%v want %v", tc.have, tc.need, got, tc.ok)
		}
	}
}

func TestMemoryMembershipIsolation(t *testing.T) {
	m := store.NewMemory("k1", "hash")
	owner, _ := m.UserByEmail("dev@recension.local")
	role, ok := m.Membership(owner.ID, "acme")
	if !ok || role != store.RoleOwner {
		t.Fatalf("bootstrap owner role=%q ok=%v", role, ok)
	}

	other, err := m.CreateUser("other@example.com", "h", "k2")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := m.Membership(other.ID, "acme"); ok {
		t.Fatal("other should not be acme member")
	}
	teams := m.ListTeamsForUser(other.ID)
	if len(teams) != 0 {
		t.Fatalf("expected 0 teams, got %d", len(teams))
	}

	acme, _ := m.TeamBySlug("acme")
	if err := m.AddMember(acme.ID, other.ID, store.RoleViewer); err != nil {
		t.Fatal(err)
	}
	if err := m.RemoveMember(acme.ID, owner.ID); err != store.ErrLastOwner {
		t.Fatalf("expected ErrLastOwner, got %v", err)
	}
}
