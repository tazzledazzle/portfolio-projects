package main

import (
	"fmt"
	"sync"
	"testing"
)

func TestCIJob_Create_SchedulesJob(t *testing.T) {
	c := NewCIJobController()
	job, err := c.Create("ci-1", "busybox:latest")
	if err != nil {
		t.Fatal(err)
	}
	if job.Kind != "CIJob" {
		t.Fatalf("kind=%q, want CIJob", job.Kind)
	}
	if job.Job == nil || job.Job.Name == "" {
		t.Fatalf("expected Job child after Create, got %+v", job)
	}
	if !job.JobScheduled {
		t.Fatalf("job_scheduled=false, want true: %+v", job)
	}
}

func TestCIJob_Reconcile_Complete(t *testing.T) {
	c := NewCIJobController()
	_, err := c.Create("ci-ok", "busybox:latest")
	if err != nil {
		t.Fatal(err)
	}
	if err := c.SetJobOutcome("ci-ok", "Succeeded"); err != nil {
		t.Fatal(err)
	}
	job, err := c.Reconcile("ci-ok")
	if err != nil {
		t.Fatal(err)
	}
	if !conditionTrue(job, "Complete") {
		t.Fatalf("Complete not True: %+v", job)
	}
	if conditionTrue(job, "Failed") {
		t.Fatalf("Failed must not be True when Complete: %+v", job)
	}
}

func TestCIJob_Reconcile_Failed(t *testing.T) {
	c := NewCIJobController()
	_, err := c.Create("ci-bad", "busybox:latest")
	if err != nil {
		t.Fatal(err)
	}
	if err := c.SetJobOutcome("ci-bad", "Failed"); err != nil {
		t.Fatal(err)
	}
	job, err := c.Reconcile("ci-bad")
	if err != nil {
		t.Fatal(err)
	}
	if !conditionTrue(job, "Failed") {
		t.Fatalf("Failed not True: %+v", job)
	}
	if conditionTrue(job, "Complete") {
		t.Fatalf("Complete must not be True when Failed: %+v", job)
	}
}

func TestCIJob_NotManagedWorkload(t *testing.T) {
	c := NewCIJobController()
	job, err := c.Create("ci-kind", "busybox:latest")
	if err != nil {
		t.Fatal(err)
	}
	if job.Kind != "CIJob" {
		t.Fatalf("kind=%q, want CIJob (not ManagedWorkload)", job.Kind)
	}
	// Primary proof is Job schedule + Complete/Failed — not finalizer_cleared.
	proof := map[string]any{
		"kind":          job.Kind,
		"job_scheduled": job.JobScheduled,
	}
	if _, ok := proof["finalizer_cleared"]; ok {
		t.Fatal("finalizer_cleared must not be a primary proof field for ENG-E-014")
	}
	if proof["kind"] != "CIJob" || proof["job_scheduled"] != true {
		t.Fatalf("unexpected proof shape: %+v", proof)
	}
}

func TestCIJob_ConcurrentReconcile(t *testing.T) {
	c := NewCIJobController()
	for i := 0; i < 32; i++ {
		id := fmt.Sprintf("ci-%d", i)
		if _, err := c.Create(id, "busybox:latest"); err != nil {
			t.Fatal(err)
		}
		outcome := "Succeeded"
		if i%2 == 1 {
			outcome = "Failed"
		}
		if err := c.SetJobOutcome(id, outcome); err != nil {
			t.Fatal(err)
		}
	}
	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.ReconcileAll()
		}()
	}
	wg.Wait()
	for i := 0; i < 32; i++ {
		id := fmt.Sprintf("ci-%d", i)
		job, ok := c.Get(id)
		if !ok {
			t.Fatalf("%s missing", id)
		}
		if i%2 == 0 {
			if !conditionTrue(job, "Complete") || conditionTrue(job, "Failed") {
				t.Fatalf("%s want Complete: %+v", id, job)
			}
		} else {
			if !conditionTrue(job, "Failed") || conditionTrue(job, "Complete") {
				t.Fatalf("%s want Failed: %+v", id, job)
			}
		}
	}
}
