package metrics

import "testing"

func TestNormalizeLabel(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		fallback string
		want     string
	}{
		{
			name:     "empty value uses fallback",
			value:    "",
			fallback: "unknown",
			want:     "unknown",
		},
		{
			name:     "whitespace value uses fallback",
			value:    "   ",
			fallback: "fallback",
			want:     "fallback",
		},
		{
			name:     "valid value is trimmed",
			value:    "  direct  ",
			fallback: "fallback",
			want:     "direct",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeLabel(tt.value, tt.fallback)
			if got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}
