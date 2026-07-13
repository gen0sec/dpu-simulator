package kind

import (
	"os"
	"testing"
)

func TestKindImageArchiveTempDir(t *testing.T) {
	t.Setenv("DPU_SIM_KIND_IMAGE_TMPDIR", "/var/tmp/kind-images")
	if got := kindImageArchiveTempDir(); got != "/var/tmp/kind-images" {
		t.Fatalf("kindImageArchiveTempDir() = %q, want %q", got, "/var/tmp/kind-images")
	}

	t.Setenv("DPU_SIM_KIND_IMAGE_TMPDIR", "")
	if got := kindImageArchiveTempDir(); got != "" {
		t.Fatalf("kindImageArchiveTempDir() = %q, want empty when unset", got)
	}

	// Ensure we read the env var at call time, not package init.
	if err := os.Unsetenv("DPU_SIM_KIND_IMAGE_TMPDIR"); err != nil {
		t.Fatalf("unsetenv: %v", err)
	}
	if got := kindImageArchiveTempDir(); got != "" {
		t.Fatalf("kindImageArchiveTempDir() = %q, want empty after unset", got)
	}
}
