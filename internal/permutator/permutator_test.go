package permutator

import "testing"

func collectAll(ch <-chan string) []string {
	var result []string
	for s := range ch {
		result = append(result, s)
	}
	return result
}

func TestCombinationCountMatchesGenerate(t *testing.T) {
	tests := []struct {
		name    string
		chunks  []string
		domains []string
		crazy   bool
	}{
		{"two_chunks_normal", []string{"john", "doe"}, []string{"gmail.com", "yahoo.com"}, false},
		{"two_chunks_crazy", []string{"john", "doe"}, []string{"gmail.com", "yahoo.com"}, true},
		{"single_chunk", []string{"alice"}, []string{"test.com"}, false},
		{"three_chunks_normal", []string{"a", "b", "c"}, []string{"x.com"}, false},
		{"three_chunks_crazy", []string{"a", "b", "c"}, []string{"x.com"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New(tt.chunks, tt.domains, tt.crazy)
			expected := p.CombinationCount()
			actual := len(collectAll(p.Generate()))
			if actual != expected {
				t.Errorf("CombinationCount()=%d but Generate() produced %d emails", expected, actual)
			}
		})
	}
}

func TestEmptyChunks(t *testing.T) {
	p := New(nil, []string{"test.com"}, false)
	if p.CombinationCount() != 0 {
		t.Error("expected 0 for empty chunks")
	}
	emails := collectAll(p.Generate())
	if len(emails) != 0 {
		t.Errorf("expected 0 emails, got %d", len(emails))
	}
}

func TestEmptyDomains(t *testing.T) {
	p := New([]string{"john"}, nil, false)
	if p.CombinationCount() != 0 {
		t.Error("expected 0 for empty domains")
	}
}

func TestSingleChunkOutput(t *testing.T) {
	p := New([]string{"alice"}, []string{"test.com"}, false)
	emails := collectAll(p.Generate())
	if len(emails) != 1 {
		t.Fatalf("expected 1 email, got %d", len(emails))
	}
	if emails[0] != "alice@test.com" {
		t.Errorf("expected alice@test.com, got %s", emails[0])
	}
}
