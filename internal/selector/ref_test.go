package selector

import "testing"

func TestNewRefResolver(t *testing.T) {
	r := NewRefResolver()
	if r == nil {
		t.Fatal("NewRefResolver returned nil")
	}
	if r.Count() != 0 {
		t.Errorf("expected 0 mappings, got %d", r.Count())
	}
}

func TestSetMappings(t *testing.T) {
	r := NewRefResolver()
	mappings := map[string]int64{
		"@e1":  1,
		"@e2":  2,
		"@e42": 42,
	}
	r.SetMappings(mappings)

	if r.Count() != 3 {
		t.Fatalf("expected 3 mappings, got %d", r.Count())
	}
	if id, ok := r.Resolve("@e1"); !ok || id != 1 {
		t.Errorf("Resolve(@e1) = (%d, %v), want (1, true)", id, ok)
	}
	if _, ok := r.Resolve("@e999"); ok {
		t.Error("Resolve(@e999) should return false")
	}
}

func TestClear(t *testing.T) {
	r := NewRefResolver()
	r.SetMappings(map[string]int64{"@e1": 1})
	r.Clear()
	if r.Count() != 0 {
		t.Errorf("expected 0 after clear, got %d", r.Count())
	}
}

func TestIsRef(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"@e1", true},
		{"@e42", true},
		{"@eabc", false},
		{"e1", false},
		{"#btn", false},
	}
	for _, tt := range tests {
		if got := IsRef(tt.input); got != tt.want {
			t.Errorf("IsRef(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestParseRef(t *testing.T) {
	n, err := ParseRef("@e42")
	if err != nil {
		t.Fatalf("ParseRef(@e42) error: %v", err)
	}
	if n != 42 {
		t.Errorf("ParseRef(@e42) = %d, want 42", n)
	}
}

func TestFormatRef(t *testing.T) {
	if got := FormatRef(1); got != "@e1" {
		t.Errorf("FormatRef(1) = %s, want @e1", got)
	}
}
