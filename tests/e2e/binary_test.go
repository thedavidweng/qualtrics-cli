package e2e

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync"
	"testing"
)

var (
	buildOnce sync.Once
	cachedBin string
	buildErr  error
)

func buildBinary(t *testing.T) string {
	t.Helper()
	buildOnce.Do(func() {
		projectRoot, err := findProjectRoot()
		if err != nil {
			buildErr = err
			return
		}
		dir, err := os.MkdirTemp("", "qualtrics-e2e-*")
		if err != nil {
			buildErr = err
			return
		}
		binName := "qualtrics"
		if runtime.GOOS == "windows" {
			binName += ".exe"
		}
		cachedBin = filepath.Join(dir, binName)
		cmd := exec.Command("go", "build", "-o", cachedBin, "./cmd/qualtrics")
		cmd.Dir = projectRoot
		if out, err := cmd.CombinedOutput(); err != nil {
			buildErr = fmt.Errorf("build failed: %v\n%s", err, out)
			return
		}
	})
	if buildErr != nil {
		t.Fatalf("binary build failed: %v", buildErr)
	}
	return cachedBin
}

func findProjectRoot() (string, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(pwd, "go.mod")); err == nil {
			return pwd, nil
		}
		parent := filepath.Dir(pwd)
		if parent == pwd {
			return "", fmt.Errorf("could not find project root (no go.mod found)")
		}
		pwd = parent
	}
}

func runCLI(t *testing.T, bin string, args ...string) (stdout, stderr string, exitCode int) {
	t.Helper()
	cmd := exec.Command(bin, args...)
	cmd.Env = append(os.Environ(),
		"QUALTRICS_CONFIG=/dev/null",
		"NO_COLOR=1",
	)
	var outBuf, errBuf strings.Builder
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err := cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			t.Fatalf("exec error: %v", err)
		}
	}
	return outBuf.String(), errBuf.String(), exitCode
}

var requiredCommands = []string{
	"auth", "completion", "definitions", "directories",
	"distributions", "doctor", "events", "raw",
	"responses", "surveys", "version",
}

var commandNameRe = regexp.MustCompile(`^\s+([a-z][a-z0-9-]*)\s+`)

func discoverCommands(t *testing.T, bin string) []string {
	t.Helper()
	stdout, _, code := runCLI(t, bin, "--help")
	if code != 0 {
		t.Fatalf("help exited with code %d", code)
	}

	seen := make(map[string]bool)
	for _, line := range strings.Split(stdout, "\n") {
		m := commandNameRe.FindStringSubmatch(line)
		if len(m) > 1 {
			cmd := m[1]
			if cmd != "help" {
				seen[cmd] = true
			}
		}
	}
	out := make([]string, 0, len(seen))
	for k := range seen {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func TestAllCommandsInHelp(t *testing.T) {
	bin := buildBinary(t)
	discovered := discoverCommands(t, bin)

	for _, required := range requiredCommands {
		found := false
		for _, d := range discovered {
			if d == required {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("required command %q not found in help: %v", required, discovered)
		}
	}
}

func TestVersionJSON(t *testing.T) {
	bin := buildBinary(t)
	stdout, _, code := runCLI(t, bin, "--json", "version")
	if code != 0 {
		t.Fatalf("version exited with code %d", code)
	}

	var env struct {
		OK   bool `json:"ok"`
		Data struct {
			Version string `json:"version"`
		} `json:"data"`
		Meta struct {
			Command string `json:"command"`
		} `json:"meta"`
	}
	if err := json.Unmarshal([]byte(stdout), &env); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, stdout)
	}
	if !env.OK || env.Meta.Command != "version" {
		t.Fatalf("unexpected envelope: %+v", env)
	}
}

func TestDoctorCommand(t *testing.T) {
	bin := buildBinary(t)
	stdout, _, code := runCLI(t, bin, "doctor")
	if code != 0 {
		t.Fatalf("doctor exited with %d", code)
	}
	if !strings.Contains(stdout, "profile:") {
		t.Fatalf("unexpected doctor output: %s", stdout)
	}
}

func TestDefinitionsBuildAndSummary(t *testing.T) {
	bin := buildBinary(t)
	tmpDir := t.TempDir()
	specFile := filepath.Join(tmpDir, "survey.md")
	qsfFile := filepath.Join(tmpDir, "survey.qsf")

	specContent := `---
title: E2E Test Survey
language: EN
---

# Block 1

## Question 1 [mc]*
- Option A
- Option B

## Question 2 [text]
`
	if err := os.WriteFile(specFile, []byte(specContent), 0o600); err != nil {
		t.Fatal(err)
	}

	stdout, stderr, code := runCLI(t, bin, "definitions", "build", specFile, "-o", qsfFile)
	if code != 0 {
		t.Fatalf("build failed: %d\nstdout: %s\nstderr: %s", code, stdout, stderr)
	}

	if _, err := os.Stat(qsfFile); err != nil {
		t.Fatalf("qsf not created: %v", err)
	}

	stdout, stderr, code = runCLI(t, bin, "--json", "definitions", "qsf", "summary", qsfFile)
	if code != 0 {
		t.Fatalf("qsf summary failed: %d\nstdout: %s\nstderr: %s", code, stdout, stderr)
	}

	var env struct {
		OK   bool `json:"ok"`
		Data struct {
			SurveyName string `json:"survey_name"`
			BlockCount int    `json:"block_count"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(stdout), &env); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if env.Data.SurveyName != "E2E Test Survey" || env.Data.BlockCount != 1 {
		t.Fatalf("unexpected summary: %+v", env)
	}
}
