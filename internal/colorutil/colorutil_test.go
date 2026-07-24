package colorutil

import "testing"

func TestIsLight(t *testing.T) {
	tests := []struct {
		color string
		want  bool
	}{
		// Hex: long form, uppercase, shorthand, and alpha variants.
		{"#ffffff", true},
		{"#FFFFFF", true},
		{"#fff", true},
		{"#ffff", true},     // #rgba shorthand, alpha ignored
		{"#ffffffff", true}, // #rrggbbaa, alpha ignored
		{"#fafafa", true},
		{"#000000", false},
		{"#102a43", false},
		{"#1a1a2e", false},
		{"#33261D", false},
		{"  #ffffff  ", true}, // surrounding whitespace
		{"#ff", false},        // too short
		// Named colors: light table entries (case-insensitive).
		{"white", true},
		{"WHITE", true},
		{"ivory", true},
		{"snow", true},
		{"beige", true},
		{"linen", true},
		{"seashell", true},
		{"floralwhite", true},
		{"ghostwhite", true},
		{"whitesmoke", true},
		{"lightyellow", true},
		{"lightgray", true},
		{"lightgrey", true},
		{"gainsboro", true},
		// Named colors: dark entries and unlisted names.
		{"black", false},
		{"navy", false},
		{"maroon", false},
		{"rebeccapurple", false}, // unlisted named color -> dark
		// rgb()/rgba() numeric forms.
		{"rgb(255, 255, 255)", true},
		{"rgb(255,250,240)", true},
		{"RGB(255, 255, 255)", true},
		{"rgba(255, 255, 255, 0.9)", true},
		{"rgb(100%, 100%, 100%)", true},
		{"rgb(255 250 240 / 0.5)", true},
		{"rgb(16, 42, 67)", false},
		{"rgba(0, 0, 0, 1)", false},
		// Unparseable input stays dark (light text is the safe default).
		{"", false},
		{"not-a-color", false},
		{"hsl(0, 0%, 100%)", false}, // hsl() is unhandled -> assumed dark
		{"rgb()", false},
		{"rgb(a, b, c)", false},
	}

	for _, tt := range tests {
		if got := IsLight(tt.color); got != tt.want {
			t.Errorf("IsLight(%q) = %v, want %v", tt.color, got, tt.want)
		}
	}
}

func TestIsSafeColor(t *testing.T) {
	tests := []struct {
		color string
		want  bool
	}{
		{"#fff", true},
		{"#ffffff", true},
		{"#ffffffff", true},
		{"  #ffffff  ", true}, // trimmed before matching
		{"white", true},
		{"rebeccapurple", true},
		{"rgb(255, 255, 255)", true},
		{"rgba(0,0,0,0.5)", true},
		{"hsl(0, 0%, 100%)", true},
		{"hsla(0, 0%, 100%, 0.5)", true},
		// Rejected: injection attempts and malformed values.
		{"", false},
		{"red; } body { color: red", false},
		{"url(x)", false},
		{"#12", false}, // fewer than 3 hex digits
		{"rgb(255, 255, 255) !important", false},
	}
	for _, tt := range tests {
		if got := IsSafeColor(tt.color); got != tt.want {
			t.Errorf("IsSafeColor(%q) = %v, want %v", tt.color, got, tt.want)
		}
	}
}
