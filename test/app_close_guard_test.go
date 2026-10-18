package test

import (
	"testing"
)

type mockCloseTester struct {
	forceQuit       bool
	runningSessions int
}

func (m *mockCloseTester) shouldPreventClose(userConfirmed bool) bool {
	if m.forceQuit {
		return false
	}
	if m.runningSessions > 0 {
		if userConfirmed {
			return false
		}
		return true // prevent close
	}
	return false
}

func TestAppCloseGuardLogic(t *testing.T) {
	// Case 1: No running sessions -> never prevent close, exit cleanly
	tester1 := &mockCloseTester{runningSessions: 0}
	if tester1.shouldPreventClose(false) {
		t.Fatalf("expected close NOT to be prevented when 0 sessions running")
	}

	// Case 2: Running sessions and user cancels -> prevent close
	tester2 := &mockCloseTester{runningSessions: 2}
	if !tester2.shouldPreventClose(false) {
		t.Fatalf("expected close to be prevented when sessions running and user cancels")
	}

	// Case 3: Running sessions and user confirms quit -> allow close
	tester3 := &mockCloseTester{runningSessions: 2}
	if tester3.shouldPreventClose(true) {
		t.Fatalf("expected close NOT to be prevented when user confirms quit")
	}

	// Case 4: Forced quit -> allow close regardless of sessions
	tester4 := &mockCloseTester{forceQuit: true, runningSessions: 5}
	if tester4.shouldPreventClose(false) {
		t.Fatalf("expected close NOT to be prevented on force quit")
	}
}
