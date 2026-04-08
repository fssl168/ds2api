package deepseek

import (
	"context"
	"testing"

	"ds2api/pow"
)

func TestPowSolverCompute(t *testing.T) {
	solver := NewPowSolver("")
	challenge := map[string]any{
		"algorithm":   "DeepSeekHashV1",
		"challenge":   "test_challenge",
		"salt":        "test_salt",
		"expire_at":   float64(1680000000),
		"difficulty":  100,
		"signature":   "sig",
		"target_path": "/chat/completions",
	}
	answer, err := solver.Compute(context.Background(), challenge)
	if err != nil {
		t.Fatalf("compute failed: %v", err)
	}
	if answer <= 0 {
		t.Fatalf("expected positive nonce, got %d", answer)
	}
}

func TestPowSolverUnsupportedAlgorithm(t *testing.T) {
	solver := NewPowSolver("")
	challenge := map[string]any{
		"algorithm": "UnknownAlgorithm",
	}
	_, err := solver.Compute(context.Background(), challenge)
	if err == nil {
		t.Fatal("expected error for unsupported algorithm")
	}
}

func TestBuildPowHeader(t *testing.T) {
	challenge := map[string]any{
		"algorithm":   "DeepSeekHashV1",
		"challenge":   "abc",
		"salt":        "salt123",
		"answer":      int64(42),
		"signature":   "test_sig",
		"target_path": "/chat/completions",
	}
	header, err := BuildPowHeader(challenge, 42)
	if err != nil {
		t.Fatalf("BuildPowHeader failed: %v", err)
	}
	if len(header) == 0 {
		t.Fatal("expected non-empty header")
	}
}

func TestNativeSolvePow(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := pow.SolvePow(ctx, "test", "salt", 1680000000, 144000)
	if err != context.Canceled {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestClientPreloadPowNoop(t *testing.T) {
	client := NewClient(nil, nil)
	if err := client.PreloadPow(context.Background()); err != nil {
		t.Fatalf("PreloadPow should be noop, got: %v", err)
	}
}
