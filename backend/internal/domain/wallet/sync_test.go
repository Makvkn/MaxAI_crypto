package wallet

import "testing"

func TestSyncStatusTransitions(t *testing.T) {
	cases := []struct {
		from  SyncStatus
		to    SyncStatus
		allow bool
	}{
		{SyncPending, SyncSyncing, true},
		{SyncPending, SyncReady, false},
		{SyncPending, SyncFailed, false},
		{SyncSyncing, SyncReady, true},
		{SyncSyncing, SyncPartial, true},
		{SyncSyncing, SyncFailed, true},
		{SyncSyncing, SyncPending, false},
		{SyncReady, SyncSyncing, true},
		{SyncFailed, SyncSyncing, true},
		{SyncReady, SyncPartial, false},
	}

	for _, tc := range cases {
		if got := tc.from.CanTransitionTo(tc.to); got != tc.allow {
			t.Errorf("%s -> %s allowed = %v, want %v", tc.from, tc.to, got, tc.allow)
		}
	}
}

func TestTerminalStatuses(t *testing.T) {
	terminal := map[SyncStatus]bool{
		SyncPending: false,
		SyncSyncing: false,
		SyncReady:   true,
		SyncPartial: true,
		SyncFailed:  true,
	}
	for status, want := range terminal {
		if got := status.IsTerminal(); got != want {
			t.Errorf("%s IsTerminal = %v, want %v", status, got, want)
		}
	}
}
