package mwb_test

import (
	"encoding/json"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"
)

// riskCategories follow frostyard/core docs/specs/risk-paths.md.
var riskCategories = []string{
	"boot chain", "signing and keys", "update verification", "auth",
	"publication", "secrets and credentials", "rulesets and settings",
	"destructive operations", "supply chain", "governance",
}

type riskPath struct {
	Glob  string `json:"glob"`
	Level string `json:"level"`
	Why   string `json:"why"`
}

func readRiskPaths(t *testing.T) []riskPath {
	t.Helper()
	data, err := os.ReadFile(".github/risk-paths.json")
	if err != nil {
		t.Fatal(err)
	}
	var list struct {
		SchemaVersion int        `json:"schema_version"`
		DefaultLevel  string     `json:"default_level"`
		Paths         []riskPath `json:"paths"`
	}
	if err := json.Unmarshal(data, &list); err != nil {
		t.Fatal(err)
	}
	if list.SchemaVersion != 2 || list.DefaultLevel != "L3" {
		t.Fatal("risk paths must use schema_version 2 and default_level L3")
	}
	if len(list.Paths) == 0 {
		t.Fatal("found 0 risk paths")
	}
	return list.Paths
}

// riskGlob anchors every glob at the root. ** matches any path; * stays
// within 1 segment. A glob without a slash never matches a nested file.
func riskGlob(glob string) *regexp.Regexp {
	var b strings.Builder
	b.WriteString("^")
	for i := 0; i < len(glob); i++ {
		switch {
		case strings.HasPrefix(glob[i:], "**"):
			if strings.Contains(glob, "/") {
				b.WriteString(".*")
			} else {
				b.WriteString("[^/]*")
			}
			i++
		case glob[i] == '*':
			b.WriteString("[^/]*")
		default:
			b.WriteString(regexp.QuoteMeta(string(glob[i])))
		}
	}
	b.WriteString("$")
	return regexp.MustCompile(b.String())
}

func riskCovered(paths []riskPath, file string) bool {
	for _, p := range paths {
		if (p.Level == "L1" || p.Level == "L2") && riskGlob(p.Glob).MatchString(file) {
			return true
		}
	}
	return false
}

func riskCategoryOK(why string) bool {
	for _, category := range riskCategories {
		if strings.HasPrefix(why, category+":") {
			return true
		}
	}
	return false
}

func TestRiskPathCategories(t *testing.T) {
	paths := readRiskPaths(t)
	for _, p := range paths {
		if !riskCategoryOK(p.Why) {
			t.Errorf("%s: why %q must start with a risk category and a colon", p.Glob, p.Why)
		}
	}
	t.Logf("checked %d risk-path categories", len(paths))
}

// These concrete files must exist and stay covered even if the marker scan
// stops finding them. Include the files under the packaging and config globs.
var riskRequired = []string{
	".github/risk-paths.json",
	".github/workflows/release.yml",
	".github/workflows/ci.yml",
	".github/workflows/pr-triage.yml",
	".github/dependabot.yml",
	"Makefile",
	".golangci.yml",
	"mise.toml",
	"mise.lock",
	"go.mod",
	"go.sum",
	".svu.yml",
	".goreleaser.yml",
	"packaging/scripts/postinstall.sh",
	"packaging/scripts/preremove.sh",
	"packaging/99-mwb.rules",
	"packaging/com.mwb.agent.plist",
	"packaging/mwb.service",
	"mwb.service",
	"internal/protocol/crypto.go",
	"internal/protocol/stream.go",
	"internal/network/client.go",
	"internal/config/config.go",
	"internal/config/config_test.go",
}

func TestRiskPathsCoverRequiredFiles(t *testing.T) {
	paths := readRiskPaths(t)
	if len(riskRequired) == 0 {
		t.Fatal("checked 0 required files")
	}
	for _, file := range riskRequired {
		if info, err := os.Stat(file); err != nil {
			t.Errorf("required file %s: %v", file, err)
		} else if !info.Mode().IsRegular() {
			t.Errorf("required file %s is not a regular file", file)
		}
		if !riskCovered(paths, file) {
			t.Errorf("%s: no L1 or L2 risk-path glob matches", file)
		}
	}
	t.Logf("checked %d required files against %d risk paths", len(riskRequired), len(paths))
}

var riskMarker = regexp.MustCompile(`os\.Remove|RemoveAll|rm -rf|rm -f|systemctl|udevadm|usermod|modprobe|loginctl|DeriveKey|NewEncryptWriter|NewDecryptReader`)

func TestRiskPathsCoverMarkerFiles(t *testing.T) {
	paths := readRiskPaths(t)
	files, err := exec.Command("git", "ls-files", "-z").Output()
	if err != nil {
		t.Fatal(err)
	}
	exemptions := map[string]string{
		"cmd/mwb/main.go":          "setup hints in error messages; they run no command",
		"internal/input/uinput.go": "setup hints in error messages; they run no command",
	}
	hits := map[string]bool{}
	checked := 0
	for _, file := range strings.Split(string(files), "\x00") {
		if file == "" || strings.HasSuffix(file, ".md") || strings.HasSuffix(file, "_test.go") || strings.HasPrefix(file, "docs/") {
			continue
		}
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		checked++
		if !riskMarker.Match(data) {
			continue
		}
		hits[file] = true
		if reason, exempt := exemptions[file]; exempt && reason != "" {
			t.Logf("exempt %s: %s", file, reason)
			continue
		}
		if !riskCovered(paths, file) {
			t.Errorf("%s has a risk marker but no L1 or L2 risk-path glob matches", file)
		}
	}
	if checked == 0 || len(hits) == 0 {
		t.Fatalf("checked %d files and found %d marker files; neither may be 0", checked, len(hits))
	}
	for file := range exemptions {
		if !hits[file] {
			t.Errorf("remove unused exemption for %s: no marker hit", file)
		}
	}
	t.Logf("scanned %d files; checked %d marker files", checked, len(hits))
}

func TestRiskPathsNegativeControl(t *testing.T) {
	paths := readRiskPaths(t)
	file := "docs/plans/x.md"
	if riskCovered(paths, file) {
		t.Fatalf("negative control %s unexpectedly has L1 or L2 coverage", file)
	}
	t.Logf("checked 1 negative control: %s is not covered", file)
}

func TestRiskCovered(t *testing.T) {
	cases := []struct {
		name  string
		paths []riskPath
		file  string
		want  bool
	}{
		{"L1", []riskPath{{Glob: "cmd/**", Level: "L1"}}, "cmd/start.go", true},
		{"L2", []riskPath{{Glob: "Makefile", Level: "L2"}}, "Makefile", true},
		{"L3", []riskPath{{Glob: "cmd/**", Level: "L3"}}, "cmd/start.go", false},
		{"L4", []riskPath{{Glob: "cmd/**", Level: "L4"}}, "cmd/start.go", false},
		{"no level", []riskPath{{Glob: "cmd/**"}}, "cmd/start.go", false},
		{"strictest wins", []riskPath{{Glob: "cmd/**", Level: "L3"}, {Glob: "cmd/start.go", Level: "L1"}}, "cmd/start.go", true},
		{"no match", []riskPath{{Glob: "internal/**", Level: "L1"}}, "cmd/start.go", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := riskCovered(tc.paths, tc.file); got != tc.want {
				t.Errorf("riskCovered(%s) = %v, want %v", tc.file, got, tc.want)
			}
		})
	}
}

func TestRiskGlob(t *testing.T) {
	cases := []struct {
		glob, file string
		want       bool
	}{
		{"packaging/**", "packaging/scripts/postinstall.sh", true},
		{"packaging/**", "packaging-extra/postinstall.sh", false},
		{"packaging/*.sh", "packaging/postinstall.sh", true},
		{"packaging/*.sh", "packaging/scripts/postinstall.sh", false},
		{"*.go", "main.go", true},
		{"*.go", "cmd/main.go", false},
		{"**", "cmd/main.go", false},
		{"Makefile", "sub/Makefile", false},
		{"a.go", "a_go", false},
		{".github/workflows/**", ".github/workflows/ci.yml", true},
	}
	for _, tc := range cases {
		if got := riskGlob(tc.glob).MatchString(tc.file); got != tc.want {
			t.Errorf("riskGlob(%q) on %q = %v, want %v", tc.glob, tc.file, got, tc.want)
		}
	}
}

func TestRiskCategoryOK(t *testing.T) {
	cases := map[string]bool{
		"auth: the handshake":          true,
		"governance: the list":         true,
		"supply chain: the tool pins":  true,
		"privileged: the helper":       false,
		"auth the handshake":           false,
		"authorization: the handshake": false,
		"":                             false,
	}
	for why, want := range cases {
		if got := riskCategoryOK(why); got != want {
			t.Errorf("riskCategoryOK(%q) = %v, want %v", why, got, want)
		}
	}
}
