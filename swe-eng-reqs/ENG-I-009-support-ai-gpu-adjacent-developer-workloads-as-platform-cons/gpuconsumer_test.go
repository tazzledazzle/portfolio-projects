package main

import (
	"bytes"
	"testing"
	"time"
)

func TestGPU_StartJob_LongRunning(t *testing.T) {
	c := NewGPUConsumer()
	job, err := c.StartJob("train-artifact", 200*time.Millisecond)
	if err != nil {
		t.Fatalf("StartJob: %v", err)
	}
	if job.ID == "" {
		t.Fatal("expected job id")
	}
	if !job.LongRunning {
		t.Fatal("expected long_running=true")
	}
	if job.Deadline.IsZero() {
		t.Fatal("expected timeout deadline")
	}
	if job.Status != "running" {
		t.Fatalf("status want running, got %s", job.Status)
	}
}

func TestGPU_UploadChunk_AccumulatesBytes(t *testing.T) {
	c := NewGPUConsumer()
	job, err := c.StartJob("upload-demo", time.Second)
	if err != nil {
		t.Fatalf("StartJob: %v", err)
	}
	out, err := c.UploadChunk(job.ID, []byte("hello-"), 0)
	if err != nil {
		t.Fatalf("UploadChunk 0: %v", err)
	}
	if out.BytesReceived != 6 {
		t.Fatalf("bytes_received want 6, got %d", out.BytesReceived)
	}
	out2, err := c.UploadChunk(job.ID, []byte("world"), 1)
	if err != nil {
		t.Fatalf("UploadChunk 1: %v", err)
	}
	if out2.BytesReceived != 11 {
		t.Fatalf("bytes_received want 11, got %d", out2.BytesReceived)
	}
	if !out2.ChunkedUpload {
		t.Fatal("expected chunked_upload=true")
	}
}

func TestGPU_Timeout_Expires(t *testing.T) {
	c := NewGPUConsumer()
	job, err := c.StartJob("timeout-job", 30*time.Millisecond)
	if err != nil {
		t.Fatalf("StartJob: %v", err)
	}
	time.Sleep(50 * time.Millisecond)
	_, err = c.UploadChunk(job.ID, []byte("late"), 0)
	if err == nil {
		t.Fatal("expected timeout error for expired job")
	}
	got, ok := c.GetJob(job.ID)
	if !ok {
		t.Fatal("job missing")
	}
	if got.Status != "timeout" && !got.TimedOut {
		t.Fatalf("expected timeout status/flag, got status=%s timed_out=%v", got.Status, got.TimedOut)
	}
}

func TestGPU_ChunkCap_RejectsOversize(t *testing.T) {
	c := NewGPUConsumer()
	job, err := c.StartJob("cap-job", time.Second)
	if err != nil {
		t.Fatalf("StartJob: %v", err)
	}
	oversize := bytes.Repeat([]byte("x"), DefaultMaxChunkBytes+1)
	_, err = c.UploadChunk(job.ID, oversize, 0)
	if err == nil {
		t.Fatal("expected oversize chunk rejection")
	}
}

func TestGPU_NotDurableWorkflow(t *testing.T) {
	c := NewGPUConsumer()
	// H-006 owns multi-step Signal API; I-009 must not expose it.
	type durableWorkflowAPI interface {
		Signal(id, action, eventID string) (any, error)
	}
	if _, ok := any(c).(durableWorkflowAPI); ok {
		t.Fatal("GPUConsumer must not expose durable Workflow Signal API (H-006 ownership)")
	}
	job, err := c.StartJob("gpu-only", time.Second)
	if err != nil {
		t.Fatalf("StartJob: %v", err)
	}
	if job.LongRunning != true {
		t.Fatal("I-009 owns long_running jobs, not durable multi-step workflows")
	}
	_, err = c.UploadChunk(job.ID, []byte("chunk"), 0)
	if err != nil {
		t.Fatalf("UploadChunk: %v", err)
	}
}
