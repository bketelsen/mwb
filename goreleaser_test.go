package mwb_test

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"
)

func TestGoReleaserVersion(t *testing.T) {
	data, err := os.ReadFile("mise.toml")
	if err != nil {
		t.Fatal(err)
	}
	pins := regexp.MustCompile(`(?m)^"github:goreleaser/goreleaser-pro" = "([0-9]+\.[0-9]+\.[0-9]+)"$`).FindAllStringSubmatch(string(data), -1)
	if len(pins) != 1 {
		t.Fatalf("expected 1 exact GoReleaser Pro pin, found %d", len(pins))
	}
	cmd := exec.Command("make", "--no-print-directory", "-s", "goreleaser-version")
	cmd.Env = goReleaserMakeEnv()
	output, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			t.Fatalf("make goreleaser-version: %v\n%s", err, exitErr.Stderr)
		}
		t.Fatalf("make goreleaser-version: %v", err)
	}
	if want := "v" + pins[0][1] + "\n"; string(output) != want {
		t.Fatalf("goreleaser-version = %q, want %q", output, want)
	}
	for _, check := range []struct {
		path, pattern string
	}{
		{"mise.lock", `(?m)^\[\[tools\."github:goreleaser/goreleaser-pro"\]\]\nversion = "` + regexp.QuoteMeta(pins[0][1]) + `"$`},
		{".github/workflows/ci.yml", `(?m)^\s*install_args:[^\n]*\bgithub:goreleaser/goreleaser-pro(?:\s|$)`},
		{"Makefile", `(?m)^ci:[^\n]*\brelease-check(?:\s|$)`},
	} {
		content, err := os.ReadFile(check.path)
		if err != nil {
			t.Fatal(err)
		}
		if !regexp.MustCompile(check.pattern).Match(content) {
			t.Errorf("%s does not wire the GoReleaser pin into CI", check.path)
		}
	}
}

func goReleaserMakeEnv() []string {
	env := []string{}
	for _, entry := range os.Environ() {
		key, _, _ := strings.Cut(entry, "=")
		if key != "MAKEFLAGS" && key != "MFLAGS" && key != "MAKELEVEL" {
			env = append(env, entry)
		}
	}
	return env
}

func TestGoReleaserVersionFailsWithoutPin(t *testing.T) {
	cmd := exec.Command("make", "--no-print-directory", "-s", "goreleaser-version", "GORELEASER_VERSION=")
	cmd.Env = goReleaserMakeEnv()
	output, err := cmd.Output()
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) || exitErr.ExitCode() == 0 {
		t.Fatalf("expected a non-zero make exit without a pin, got %v", err)
	}
	if len(output) != 0 {
		t.Fatalf("expected empty stdout without a pin, got %q; stderr: %s", output, exitErr.Stderr)
	}
}

// checkGoReleaserSteps checks inputs only within each action step's with block.
func checkGoReleaserSteps(workflow string) (int, error) {
	lines := strings.Split(workflow, "\n")
	indent := func(line string) int { return len(line) - len(strings.TrimLeft(line, " ")) }
	count := 0
	for start, line := range lines {
		if !strings.HasPrefix(strings.TrimSpace(line), "- ") {
			continue
		}
		level := indent(line)
		end := start + 1
		for end < len(lines) {
			if strings.HasPrefix(strings.TrimSpace(lines[end]), "- ") && indent(lines[end]) <= level {
				break
			}
			end++
		}
		action := false
		for i := start; i < end; i++ {
			field := strings.TrimSpace(lines[i])
			if i == start {
				field = strings.TrimPrefix(field, "- ")
			} else if indent(lines[i]) != level+2 {
				continue
			}
			if strings.HasPrefix(field, "uses: goreleaser/goreleaser-action@") {
				action = true
			}
		}
		if !action {
			continue
		}
		count++
		distribution, version := "", ""
		for i := start + 1; i < end; i++ {
			if indent(lines[i]) != level+2 || strings.TrimSpace(lines[i]) != "with:" {
				continue
			}
			for j := i + 1; j < end; j++ {
				field := strings.TrimSpace(lines[j])
				if field == "" || strings.HasPrefix(field, "#") {
					continue
				}
				if indent(lines[j]) <= level+2 {
					break
				}
				if indent(lines[j]) != level+4 {
					continue
				}
				key, value, ok := strings.Cut(field, ":")
				if !ok {
					continue
				}
				switch key {
				case "distribution":
					distribution = strings.TrimSpace(value)
				case "version":
					version = strings.TrimSpace(value)
				}
			}
		}
		if distribution != "goreleaser-pro" {
			return count, fmt.Errorf("GoReleaser step %d must use distribution: goreleaser-pro", count)
		}
		if version != "${{ steps.goreleaser-pin.outputs.version }}" {
			return count, fmt.Errorf("GoReleaser step %d must use the shared version pin, found %q", count, version)
		}
	}
	if count == 0 {
		return 0, fmt.Errorf("found 0 GoReleaser steps")
	}
	return count, nil
}

func TestReleaseUsesGoReleaserPin(t *testing.T) {
	data, err := os.ReadFile(".github/workflows/release.yml")
	if err != nil {
		t.Fatal(err)
	}
	count, err := checkGoReleaserSteps(string(data))
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Fatalf("expected 2 GoReleaser steps, found %d", count)
	}
	t.Logf("checked %d GoReleaser steps", count)
	jobs := regexp.MustCompile(`(?m)^  (?:release|merge):\n`).Split(string(data), -1)
	if len(jobs) != 3 {
		t.Fatalf("expected release and merge jobs, found %d", len(jobs)-1)
	}
	for _, job := range jobs[1:] {
		readPin := "      - name: Read the GoReleaser pin\n" +
			"        id: goreleaser-pin\n" +
			"        run: |\n" +
			"          version=\"$(make --no-print-directory -s goreleaser-version)\"\n" +
			"          echo \"version=$version\" >> \"$GITHUB_OUTPUT\"\n"
		if strings.Count(job, readPin) != 1 {
			t.Error("each release job must read the pin with a separate make assignment")
		}
		action := "        uses: goreleaser/goreleaser-action@"
		if strings.Count(job, action) != 1 {
			t.Fatal("each release job must have 1 GoReleaser action")
		}
		if strings.Index(job, readPin) > strings.Index(job, action) {
			t.Error("read the pin before the GoReleaser action")
		}
	}
}

func TestCheckGoReleaserStepsRejects(t *testing.T) {
	valid := "steps:\n  - uses: goreleaser/goreleaser-action@abc\n    with:\n      distribution: goreleaser-pro\n      version: ${{ steps.goreleaser-pin.outputs.version }}\n"
	for _, tc := range []struct {
		name, workflow string
		wantCount      int
		wantErr        bool
	}{
		{"floating version", strings.ReplaceAll(valid, "${{ steps.goreleaser-pin.outputs.version }}", `"~> v2"`), 1, true},
		{"no steps", "steps:\n  - run: true\n", 0, true},
		{"missing distribution", strings.ReplaceAll(valid, "      distribution: goreleaser-pro\n", ""), 1, true},
		{"valid", valid, 1, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			count, err := checkGoReleaserSteps(tc.workflow)
			if count != tc.wantCount || (err != nil) != tc.wantErr {
				t.Fatalf("checkGoReleaserSteps = (%d, %v), want (%d, error=%v)", count, err, tc.wantCount, tc.wantErr)
			}
		})
	}
}
