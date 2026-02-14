package service

import "testing"

func TestShouldTriggerRule(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		mode      string
		threshold int32
		count     int64
		want      bool
	}{
		{
			name:      "at_triggers_when_count_equals_threshold",
			mode:      "at",
			threshold: 3,
			count:     3,
			want:      true,
		},
		{
			name:      "every_does_not_trigger_at_zero",
			mode:      "every",
			threshold: 3,
			count:     0,
			want:      false,
		},
		{
			name:      "every_triggers_on_multiple",
			mode:      "every",
			threshold: 3,
			count:     6,
			want:      true,
		},
		{
			name:      "every_does_not_trigger_on_non_multiple",
			mode:      "every",
			threshold: 3,
			count:     5,
			want:      false,
		},
		{
			name:      "after_triggers_when_count_is_greater",
			mode:      "after",
			threshold: 3,
			count:     4,
			want:      true,
		},
		{
			name:      "invalid_threshold_never_triggers",
			mode:      "every",
			threshold: 0,
			count:     10,
			want:      false,
		},
		{
			name:      "invalid_mode_never_triggers",
			mode:      "unknown",
			threshold: 3,
			count:     3,
			want:      false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := shouldTriggerRule(tt.mode, tt.threshold, tt.count)
			if got != tt.want {
				t.Fatalf("shouldTriggerRule(%q, %d, %d) = %v, want %v", tt.mode, tt.threshold, tt.count, got, tt.want)
			}
		})
	}
}
