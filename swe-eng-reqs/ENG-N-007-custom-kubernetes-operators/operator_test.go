package main

import (
	"fmt"
	"sync"
	"testing"
)

func TestReconcile_SetsReady(t *testing.T) {
	controller := NewController()
	if _, err := controller.Create("payments", 3, "example/payments:v1"); err != nil {
		t.Fatalf("Create: %v", err)
	}

	workload, err := controller.Reconcile("payments")
	if err != nil {
		t.Fatalf("Reconcile: %v", err)
	}
	if !conditionTrue(workload, "Ready") {
		t.Fatalf("expected Ready=True, got %#v", workload.Conditions)
	}
}

func TestFinalizer_DeleteRetainsUntilFinalize(t *testing.T) {
	controller := NewController()
	if _, err := controller.Create("worker", 2, "example/worker:v1"); err != nil {
		t.Fatalf("Create: %v", err)
	}

	deleting, err := controller.Delete("worker")
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if !deleting.Deleting || len(deleting.Finalizers) == 0 {
		t.Fatalf("expected deleting workload with finalizer, got %#v", deleting)
	}
	if _, ok := controller.Get("worker"); !ok {
		t.Fatal("workload disappeared before finalization")
	}

	finalized, err := controller.Finalize("worker")
	if err != nil {
		t.Fatalf("Finalize: %v", err)
	}
	if len(finalized.Finalizers) != 0 {
		t.Fatalf("expected finalizers cleared, got %#v", finalized.Finalizers)
	}
	if _, ok := controller.Get("worker"); ok {
		t.Fatal("workload still present after finalization")
	}
}

func TestFinalizer_WithoutFinalize_StillPresent(t *testing.T) {
	controller := NewController()
	if _, err := controller.Create("api", 1, "example/api:v1"); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if _, err := controller.Delete("api"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, ok := controller.Get("api"); !ok {
		t.Fatal("delete must retain workload until Finalize")
	}
}

func TestReconcile_Idempotent(t *testing.T) {
	controller := NewController()
	if _, err := controller.Create("frontend", 2, "example/frontend:v1"); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if _, err := controller.Reconcile("frontend"); err != nil {
		t.Fatalf("first Reconcile: %v", err)
	}
	workload, err := controller.Reconcile("frontend")
	if err != nil {
		t.Fatalf("second Reconcile: %v", err)
	}
	if !conditionTrue(workload, "Ready") {
		t.Fatalf("expected Ready=True, got %#v", workload.Conditions)
	}
	if len(workload.Conditions) != 1 {
		t.Fatalf("expected one upserted condition, got %d", len(workload.Conditions))
	}
}

func TestOperator_ConcurrentReconcile(t *testing.T) {
	controller := NewController()
	for i := 0; i < 20; i++ {
		id := fmt.Sprintf("workload-%d", i)
		if _, err := controller.Create(id, 1, "example/app:v1"); err != nil {
			t.Fatalf("Create(%s): %v", id, err)
		}
		if i%2 == 1 {
			if _, err := controller.Delete(id); err != nil {
				t.Fatalf("Delete(%s): %v", id, err)
			}
		}
	}

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			id := fmt.Sprintf("workload-%d", i)
			if i%2 == 0 {
				if _, err := controller.Reconcile(id); err != nil {
					t.Errorf("Reconcile(%s): %v", id, err)
				}
				return
			}
			if _, err := controller.Finalize(id); err != nil {
				t.Errorf("Finalize(%s): %v", id, err)
			}
		}()
	}
	wg.Wait()

	for i := 0; i < 20; i++ {
		id := fmt.Sprintf("workload-%d", i)
		workload, ok := controller.Get(id)
		if i%2 == 1 {
			if ok {
				t.Errorf("%s remains after Finalize", id)
			}
			continue
		}
		if !ok || !conditionTrue(workload, "Ready") {
			t.Errorf("%s was not reconciled Ready", id)
		}
	}
}
