package service

import "testing"

func TestShouldTriggerRule(t *testing.T) {
	tests := []struct {
		name      string
		mode      string
		threshold int32
		count     int64
		want      bool
	}{
		{name: "at true", mode: "at", threshold: 3, count: 3, want: true},
		{name: "at false", mode: "at", threshold: 3, count: 2, want: false},
		{name: "every true", mode: "every", threshold: 2, count: 4, want: true},
		{name: "every false when zero", mode: "every", threshold: 2, count: 0, want: false},
		{name: "after true", mode: "after", threshold: 3, count: 4, want: true},
		{name: "after false equal", mode: "after", threshold: 3, count: 3, want: false},
		{name: "invalid mode", mode: "unknown", threshold: 3, count: 3, want: false},
		{name: "invalid threshold", mode: "at", threshold: 0, count: 3, want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := shouldTriggerRule(tc.mode, tc.threshold, tc.count)
			if got != tc.want {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
		})
	}
}
