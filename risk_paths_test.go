package mwb_test

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
	for _, p := range list.Paths {
		if _, err := riskGlob(p.Glob); err != nil {
			t.Fatalf("invalid risk-path glob %q: %v", p.Glob, err)
		}
	}
	return list.Paths
}

// riskGlob anchors globs at the root, except ** alone matches every path.
// **/ matches 0 or more segments; a trailing /** matches paths below a folder.
// This gate refuses question marks, character classes, negation and braces.
func riskGlob(glob string) (*regexp.Regexp, error) {
	if strings.ContainsAny(glob, `?[!{}\`) {
		return nil, fmt.Errorf("unsupported glob syntax")
	}
	segments := strings.Split(glob, "/")
	for _, segment := range segments {
		if segment == "" || segment == "." || segment == ".." {
			return nil, fmt.Errorf("glob must have root-relative, non-empty segments")
		}
		if segment != "**" && strings.Contains(segment, "**") {
			return nil, fmt.Errorf("** must be a whole segment")
		}
	}
	var b strings.Builder
	b.WriteString("^")
	for i := 0; i < len(segments); i++ {
		segment := segments[i]
		if segment == "**" {
			if i == len(segments)-1 {
				b.WriteString(".*")
			} else {
				b.WriteString("(?:[^/]+/)*")
			}
			continue
		}
		for _, char := range segment {
			if char == '*' {
				b.WriteString("[^/]*")
			} else {
				b.WriteString(regexp.QuoteMeta(string(char)))
			}
		}
		if i < len(segments)-1 {
			b.WriteString("/")
		}
	}
	b.WriteString("$")
	return regexp.Compile(b.String())
}

func riskCovered(paths []riskPath, file string) bool {
	for _, p := range paths {
		if p.Level == "L1" || p.Level == "L2" {
			pattern, err := riskGlob(p.Glob)
			if err == nil && pattern.MatchString(file) {
				return true
			}
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

func riskScanSkip(file string) (bool, error) {
	if file == "" || strings.HasSuffix(file, ".md") || strings.HasSuffix(file, "_test.go") || strings.HasPrefix(file, "docs/") {
		return true, nil
	}
	info, err := os.Lstat(file)
	if err != nil {
		return false, err
	}
	// Git stores a symlink's target path, not its content; the tracked target is scanned under its own name.
	return info.Mode()&os.ModeSymlink != 0, nil
}

func TestRiskMarkerScanSkipsSymlinks(t *testing.T) {
	dir := t.TempDir()
	real := filepath.Join(dir, "real.txt")
	if err := os.WriteFile(real, []byte("modprobe"), 0o600); err != nil {
		t.Fatal(err)
	}
	alias := filepath.Join(dir, "alias")
	if err := os.Symlink(real, alias); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		file string
		want bool
	}{
		{alias, true},
		{real, false},
	} {
		if skip, err := riskScanSkip(tc.file); err != nil {
			t.Fatal(err)
		} else if skip != tc.want {
			t.Errorf("riskScanSkip(%q) = %v, want %v", tc.file, skip, tc.want)
		}
	}
}

func TestRiskPathsCoverMarkerFiles(t *testing.T) {
	inside, err := exec.Command("git", "rev-parse", "--is-inside-work-tree").CombinedOutput()
	if err != nil || strings.TrimSpace(string(inside)) != "true" {
		if os.Getenv("CI") == "" {
			t.Skip("source tarball: no git work tree is available for the tracked-file marker scan")
		}
		t.Fatalf("CI requires a git work tree for the marker scan: %v; %s", err, inside)
	}
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
	checked, symlinks := 0, 0
	for _, file := range strings.Split(string(files), "\x00") {
		skip, err := riskScanSkip(file)
		if err != nil {
			t.Fatal(err)
		}
		if skip {
			if info, err := os.Lstat(file); err == nil && info.Mode()&os.ModeSymlink != 0 {
				symlinks++
			}
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
	t.Logf("scanned %d files; skipped %d symlinks; checked %d marker files", checked, symlinks, len(hits))
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
		{"*.go", "a.go", true},
		{"*.go", "cmd/a.go", false},
		{"**", "cmd/mwb/main.go", true},
		{"**/x.go", "x.go", true},
		{"**/x.go", "a/b/x.go", true},
		{"internal/**", "internal", false},
		{"Makefile", "sub/Makefile", false},
		{"a.go", "a_go", false},
		{".github/workflows/**", ".github/workflows/ci.yml", true},
	}
	for _, tc := range cases {
		pattern, err := riskGlob(tc.glob)
		if err != nil {
			t.Fatalf("riskGlob(%q): %v", tc.glob, err)
		}
		if got := pattern.MatchString(tc.file); got != tc.want {
			t.Errorf("riskGlob(%q) on %q = %v, want %v", tc.glob, tc.file, got, tc.want)
		}
	}
	for _, glob := range []string{"a?.go", "[ab].go", "!a.go", "{a,b}.go", "/a.go", "./a.go", "a//b", "a/../b", "a**/b"} {
		if _, err := riskGlob(glob); err == nil {
			t.Errorf("riskGlob(%q) must refuse unsupported syntax", glob)
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
