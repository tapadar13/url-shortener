package auth

import "testing"

func TestRepositoryErrorsAreDistinct(t *testing.T) {
	if ErrUserNotFound == ErrEmailTaken {
		t.Fatal("expected distinct repository errors")
	}
}
