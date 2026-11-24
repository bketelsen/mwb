package mwb_test

import (
	"os/exec"
	"strings"
	"testing"
)

// makeDryRun prints the commands of a make target without running them.
// GNU make still runs the $(MAKE) lines, with -n passed on, so the output
// covers the sub-makes too.
func makeDryRun(t *testing.T, target string) string {
	t.Helper()
	if _, err := exec.LookPath("make"); err != nil {
		t.Skipf("make is not on PATH: %v", err)
	}
	cmd := exec.Command("make", "--no-print-directory", "-n", target)
	cmd.Env = goReleaserMakeEnv()
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("make -n %s: %v\n%s", target, err, output)
	}
	return string(output)
}

// make bump must not need GoReleaser Pro on the maintainer's machine: the
// release runs it in CI (release.yml), and ci.yml runs release-check on main.
func TestBumpNeedsNoGoReleaser(t *testing.T) {
	bump := makeDryRun(t, "bump")
	for _, unwanted := range []string{"goreleaser --version", "goreleaser check"} {
		if strings.Contains(bump, unwanted) {
			t.Errorf("make bump runs %q:\n%s", unwanted, bump)
		}
	}
	for _, want := range []string{"svu --version", "git fetch --quiet --tags origin main", "golangci-lint run", "go test -race ./...", "build/mwb-linux-arm64", "svu next"} {
		if !strings.Contains(bump, want) {
			t.Errorf("make bump does not run %q:\n%s", want, bump)
		}
	}
	if ci := makeDryRun(t, "ci"); !strings.Contains(ci, "goreleaser check") {
		t.Errorf("make ci no longer runs goreleaser check:\n%s", ci)
	}
}
