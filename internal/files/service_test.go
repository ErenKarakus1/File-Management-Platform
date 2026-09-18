package files

import "testing"

func TestSanitizeOriginalName(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "base name", in: "report.pdf", want: "report.pdf"},
		{name: "windows path", in: `C:\temp\report.pdf`, want: "report.pdf"},
		{name: "unix path", in: "/tmp/report.pdf", want: "report.pdf"},
		{name: "blank", in: "   ", want: "upload"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sanitizeOriginalName(tt.in); got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}
