package importurl

import "testing"

func TestNormalizeSourceURL(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{
			in:   "https://usu.instructure.com/courses/807136/assignments/5218393?submitting=1",
			want: "https://usu.instructure.com/courses/807136/assignments/5218393",
		},
		{
			in:   "https://school.instructure.com/courses/1/assignments/2#rubric",
			want: "https://school.instructure.com/courses/1/assignments/2",
		},
		{
			in:   "  https://school.instructure.com/courses/1/assignments/2/  ",
			want: "https://school.instructure.com/courses/1/assignments/2",
		},
	}

	for _, tc := range tests {
		if got := NormalizeSourceURL(tc.in); got != tc.want {
			t.Fatalf("NormalizeSourceURL(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
