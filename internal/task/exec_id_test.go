package task

import "testing"

func TestGenerateExecID_Length(t *testing.T) {
	id := generateExecID()
	if len(id) != 8 {
		t.Errorf("len = %d, want 8", len(id))
	}
}

func TestGenerateExecID_Charset(t *testing.T) {
	id := generateExecID()
	for _, c := range id {
		if !isBase36(c) {
			t.Errorf("invalid char %q in %q", c, id)
		}
	}
}

func TestGenerateExecID_Unique(t *testing.T) {
	const n = 10000
	seen := make(map[string]bool, n)
	for i := 0; i < n; i++ {
		id := generateExecID()
		if seen[id] {
			t.Fatalf("duplicate id: %s", id)
		}
		seen[id] = true
	}
}

func isBase36(c rune) bool {
	return (c >= '0' && c <= '9') || (c >= 'A' && c <= 'Z')
}
