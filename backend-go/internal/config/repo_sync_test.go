package config

import (
	"testing"

	"github.com/spf13/viper"
)

// The difference this pins is the whole reason the helper exists: an UNSET
// variable means "use the default schedule", while an explicit 0 means "no
// schedule". Treating them the same would either ignore an operator who wants the
// sweep gone, or silently leave the feature off for everyone who never heard of
// the variable.
func TestRepoSyncIntervalHours(t *testing.T) {
	tests := []struct {
		name  string
		set   bool
		value int
		want  int
	}{
		{name: "unset means the 6h default", set: false, want: 6},
		{name: "explicit zero turns the schedule off", set: true, value: 0, want: 0},
		{name: "explicit value is honoured", set: true, value: 12, want: 12},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			v := viper.New()
			if tc.set {
				v.Set("REPO_SYNC_INTERVAL_HOURS", tc.value)
			}
			if got := repoSyncIntervalHours(v); got != tc.want {
				t.Fatalf("repoSyncIntervalHours = %d, want %d", got, tc.want)
			}
		})
	}
}
