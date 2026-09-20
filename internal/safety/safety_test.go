package safety

import "testing"

func TestCheck(t *testing.T) {
	cases := []struct {
		name      string
		tier      OperationTier
		readOnly  bool
		dryRun    bool
		confirmed bool
		wantCode  bool
	}{
		{"read always passes", TierRead, true, false, false, false},
		{"mutation needs confirm", TierMutation, false, false, false, true},
		{"mutation with confirm passes", TierMutation, false, false, true, false},
		{"destructive with dry-run passes", TierDestructive, false, true, false, false},
		{"read-only blocks mutation", TierMutation, true, false, true, true},
		{"remote action needs confirm", TierRemoteAction, false, false, false, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := Check(c.tier, c.readOnly, c.dryRun, c.confirmed)
			if c.wantCode && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !c.wantCode && err != nil {
				t.Fatalf("expected pass, got %v", err)
			}
		})
	}
}
