package enumerator

import "testing"

func TestDetectHashType(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"437e4dc6d001f2519bc9e7a6b6412923", "MD5"},
		{"ABCDEF1234567890abcdef1234567890", "MD5"},
		{"e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", "SHA256"},
		{"username", ""},
		{"abc", ""},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := DetectHashType(tt.input)
			if got != tt.want {
				t.Errorf("DetectHashType(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
