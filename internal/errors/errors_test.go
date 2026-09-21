package errors

import "testing"

func TestExitCodes(t *testing.T) {
	cases := []struct {
		code Code
		want int
	}{
		{AuthRequired, 3},
		{AuthTokenInvalid, 3},
		{ReadOnlyViolation, 4},
		{RateLimited, 5},
		{NetworkUnreachable, 5},
		{NetworkTimeout, 5},
		{APIError, 6},
		{APISchemaChanged, 6},
		{APIAccessForbidden, 6},
		{ResourceNotFound, 6},
		{ValidationFailed, 7},
		{ConfirmationRequired, 10},
		{InvalidArguments, 2},
		{InternalError, 1},
	}
	for _, c := range cases {
		e := New(c.code, "m", CatInternal, false, nil)
		if got := e.ExitCode(); got != c.want {
			t.Errorf("%s exit = %d, want %d", c.code, got, c.want)
		}
	}
}
