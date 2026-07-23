package main

import (
	"strings"
	"testing"
)

func TestRenderMethodShowsOnlyFullSourceBlock(t *testing.T) {
	method := &Method{
		baseMember: baseMember{Name: "EndTransaction", Scope: ScopePublic},
		Signature:  "Sub EndTransaction()",
		Source: "Sub EndTransaction()\n" +
			"  Self.CommitTransaction\n" +
			"End Sub",
	}
	var output strings.Builder

	renderMethod(&output, method, &renderCtx{})
	rendered := output.String()

	if strings.Contains(rendered, "<details") || strings.Contains(rendered, "<summary>Source</summary>") {
		t.Fatalf("method source still uses a disclosure wrapper:\n%s", rendered)
	}
	if strings.Count(rendered, "Sub EndTransaction()") != 1 {
		t.Fatalf("method signature should occur once inside full source:\n%s", rendered)
	}
	for _, line := range []string{"Self.CommitTransaction", "End Sub"} {
		if !strings.Contains(rendered, line) {
			t.Fatalf("full source is missing %q:\n%s", line, rendered)
		}
	}
}
