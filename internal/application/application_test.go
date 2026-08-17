package application

import "testing"

// TestStoreAdd check the new(first) application in memory
func TestStoreAdd(t *testing.T) {
	s := NewStore()
	got := s.Add(Application{Client: "Alina", Amount: 1000, Term: 12})

	if got.ID != 1 {
		t.Errorf("expected ID 1, got %d", got.ID)
	}

	if got.Status != "new" {
		t.Errorf("expected status new, got %s", got.Status)
	}
}