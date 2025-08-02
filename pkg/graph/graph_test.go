package graph

import "testing"

func TestGraphRelationships(t *testing.T) {
	g := New()
	g.AddRelation("user:alice", "group:managers")
	g.AddRelation("group:managers", "resource:db")

	if !g.HasPath("user:alice", "group:managers") {
		t.Fatalf("expected user to be in group")
	}
	if !g.HasPath("group:managers", "resource:db") {
		t.Fatalf("expected group to relate to resource")
	}
	if !g.HasPath("user:alice", "resource:db") {
		t.Fatalf("expected path from user to resource")
	}
}
