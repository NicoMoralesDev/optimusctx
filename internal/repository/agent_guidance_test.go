package repository

import (
	"strings"
	"testing"
)

func TestMergeManagedGuidanceAppendsWithoutRewritingUserContent(t *testing.T) {
	existing := "# User rules\n\nkeep trailing spaces here   \nDo not rewrite this section.\n"

	merged, err := MergeManagedGuidance([]byte(existing), RenderCodexGuidanceBlock())
	if err != nil {
		t.Fatalf("MergeManagedGuidance() error = %v", err)
	}

	if !strings.HasPrefix(merged, existing) {
		t.Fatalf("merged content should preserve existing prefix exactly:\n%q", merged)
	}
	if strings.Count(merged, managedGuidanceBegin) != 1 {
		t.Fatalf("managed guidance block count = %d, want 1", strings.Count(merged, managedGuidanceBegin))
	}
	if !strings.HasSuffix(merged, managedGuidanceEnd+"\n") {
		t.Fatalf("merged content should end with managed block:\n%q", merged)
	}
}

func TestMergeManagedGuidanceReplacesOnlyManagedBlock(t *testing.T) {
	existing := strings.Join([]string{
		"# User rules",
		"",
		"keep this intro",
		managedGuidanceBegin,
		"old guidance",
		managedGuidanceEnd,
		"",
		"keep this footer exactly",
		"",
	}, "\n")

	merged, err := MergeManagedGuidance([]byte(existing), RenderCodexGuidanceBlock())
	if err != nil {
		t.Fatalf("MergeManagedGuidance() error = %v", err)
	}

	if !strings.HasPrefix(merged, "# User rules\n\nkeep this intro\n") {
		t.Fatalf("merged content rewrote prefix:\n%q", merged)
	}
	if !strings.HasSuffix(merged, "\nkeep this footer exactly\n") {
		t.Fatalf("merged content rewrote suffix:\n%q", merged)
	}
	if strings.Count(merged, managedGuidanceBegin) != 1 {
		t.Fatalf("managed guidance block count = %d, want 1", strings.Count(merged, managedGuidanceBegin))
	}
}

func TestMergeManagedGuidancePreservesTrailingNewlineStyleWhenAppending(t *testing.T) {
	existing := "# User rules without trailing newline"

	merged, err := MergeManagedGuidance([]byte(existing), RenderCodexGuidanceBlock())
	if err != nil {
		t.Fatalf("MergeManagedGuidance() error = %v", err)
	}

	if !strings.HasPrefix(merged, existing+"\n\n") {
		t.Fatalf("merged content should append after a blank line:\n%q", merged)
	}
}
