package cli_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/mineiros-io/terrastack/cmd/terrastack/cli"
	"github.com/mineiros-io/terrastack/test"
	"github.com/mineiros-io/terrastack/test/sandbox"
)

func TestBug25(t *testing.T) {
	// bug: https://github.com/mineiros-io/terrastack/issues/25

	const (
		mod1 = "1"
		mod2 = "2"
	)

	s := sandbox.New(t)

	mod1MainTf := s.CreateModule(mod1).CreateFile("main.tf", "# module 1")
	s.CreateModule(mod2).CreateFile("main.tf", "# module 2")

	stack1 := s.CreateStack("stack-1")
	stack2 := s.CreateStack("stack-2")
	stack3 := s.CreateStack("stack-3")

	stack1.CreateFile("main.tf", `
module "mod1" {
source = "%s"
}`, stack1.ModSource(mod1))

	stack2.CreateFile("main.tf", `
module "mod2" {
source = "%s"
}`, stack2.ModSource(mod2))

	stack3.CreateFile("main.tf", "# no module")

	cli := newCLI(t)
	cli.run("init", stack1.Path(), stack2.Path(), stack3.Path())

	git := s.Git()
	git.CommitAll("first commit")

	assertRun(t, cli.run("list", s.BaseDir(), "--changed"), runResult{})

	git.CheckoutNew("change-the-module-1")

	mod1MainTf.Write("# changed")

	git.CommitAll("module 1 changed")

	want := stack1.Path() + "\n"
	assertRun(t, cli.run(
		"list", s.BaseDir(), "--changed"),
		runResult{Stdout: want},
	)
}

func TestListAndRunChangedStack(t *testing.T) {
	const (
		mainTfFileName = "main.tf"
		mainTfContents = "# change is the eternal truth of the universe"
	)

	s := sandbox.New(t)

	stack := s.CreateStack("stack")
	stackMainTf := stack.CreateFile(mainTfFileName, "# some code")

	cli := newCLI(t)
	cli.run("init", stack.Path())

	git := s.Git()
	git.CommitAll("first commit")

	assertRun(t, cli.run("list", s.BaseDir(), "--changed"), runResult{})

	git.CheckoutNew("change-stack")

	stackMainTf.Write(mainTfContents)
	git.CommitAll("stack changed")

	wantList := stack.Path() + "\n"
	assertRun(t, cli.run("list", s.BaseDir(), "--changed"), runResult{Stdout: wantList})

	cat := test.LookPath(t, "cat")
	wantRun := fmt.Sprintf(
		"Running on changed stacks:\n[%s] running %s %s\n%s",
		stack.Path(),
		cat,
		mainTfFileName,
		mainTfContents,
	)

	assertRun(t, cli.run(
		"run",
		"--basedir",
		s.BaseDir(),
		"--changed",
		cat,
		mainTfFileName,
	), runResult{Stdout: wantRun})
}

func TestNoArgsProvidesBasicHelp(t *testing.T) {
	cli := newCLI(t)
	cli.run("--help")
	help := cli.run("--help")
	assertRun(t, cli.run(), runResult{Stdout: help.Stdout})
}

type runResult struct {
	Cmd    string
	Stdout string
	Stderr string
}

type tscli struct {
	t *testing.T
}

func newCLI(t *testing.T) tscli {
	return tscli{t: t}
}

func (ts tscli) run(args ...string) runResult {
	ts.t.Helper()

	stdin := &bytes.Buffer{}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	if err := cli.Run(args, stdin, stdout, stderr); err != nil {
		ts.t.Fatalf(
			"cli.Run(args=%v) error=%q stdout=%q stderr=%q",
			args,
			err,
			stdout.String(),
			stderr.String(),
		)
	}

	return runResult{
		Cmd:    strings.Join(args, " "),
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}
}

func assertRun(t *testing.T, got runResult, want runResult) {
	t.Helper()

	if got.Stdout != want.Stdout {
		t.Errorf("%q stdout=%q, wanted=%q", got.Cmd, got.Stdout, want.Stdout)
	}

	if got.Stderr != want.Stderr {
		t.Errorf("%q stderr=%q, wanted=%q", got.Cmd, got.Stderr, want.Stderr)
	}
}
