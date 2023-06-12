package easyworker

import (
	"context"
	"testing"
)

func TestSupManGet(t *testing.T) {
	sup := NewSupervisorWithContext(context.Background())

	sup.NewChild(NO_RESTART, simpleLoopWithContext, 3)

	if GetSupervisor(sup.id) == nil {
		t.Error("cannot get supervisor by id")
	}

	sup.Stop()
}

func TestSupManRemove(t *testing.T) {
	sup := NewSupervisorWithContext(context.Background())

	sup.NewChild(NO_RESTART, simpleLoopWithContext, 3)

	sup.Stop()

	RemoveSupervisorById(sup.id)

	if GetSupervisor(sup.id) != nil {
		t.Error("cannot get supervisor by id")
	}
}
