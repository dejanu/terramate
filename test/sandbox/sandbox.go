// Package sandbox provides an easy way to setup isolated terrastack projects
// that can be used on testing, acting like sandboxes.
//
// It helps with:
//
// - git initialization/operations
// - Terraform module creation
// - Terrastack stack creation
package sandbox

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/madlambda/spells/assert"
	"github.com/mineiros-io/terrastack"
	"github.com/mineiros-io/terrastack/test"
)

// S is a full sandbox with its own base dir that is an initialized git repo for
// test purposes.
type S struct {
	t       *testing.T
	git     Git
	basedir string
}

// DirEntry represents a directory and can be used to create files inside the
// directory.
type DirEntry struct {
	t       *testing.T
	abspath string
	relpath string
}

// StackEntry represents a directory that's also a stack.
// It extends a DirEntry with stack specific functionality.
type StackEntry struct {
	DirEntry
}

// FileEntry represents a file and can be used to manipulate the file contents.
// It is optimized for reading/writing all contents, not stream programming
// (io.Reader/io.Writer).
// It has limited usefulness but it is easier to work with for testing.
type FileEntry struct {
	t    *testing.T
	path string
}

// New creates a new test sandbox.
//
// It is a programming error to use a test env created with a *testing.T other
// than the one of the test using the test env, for a new test/sub-test always
// create a new test env for it.
func New(t *testing.T) S {
	t.Helper()

	basedir := t.TempDir()
	git := NewGit(t, basedir)
	git.Init()
	return S{
		t:       t,
		git:     git,
		basedir: basedir,
	}
}

// BuildTree builds a tree layout based on the layout specification, defined
// below:
// Each string in the slice represents a filesystem operation, and each
// operation has the format below:
//   <kind>:<relative path>[:data]
// Where kind is one of the below:
//   "d" for directory creation.
//   "s" for initialized stacks.
//   "f" for file creation
// The data field is optional and only used with "f" for the file content.
//
// This is an internal mini-lang used to simplify testcases, so it expects well
// formed layout specification.
func (s S) BuildTree(layout []string) {
	t := s.t
	t.Helper()

	for _, spec := range layout {
		switch spec[0] {
		case 'd':
			test.MkdirAll(t, filepath.Join(s.basedir, spec[2:]))
		case 's':
			s.CreateStack(spec[2:])
		case 'f':
			tmp := spec[2:]
			index := strings.IndexByte(tmp, ':')
			file := tmp[0:index]
			content := tmp[index+1:]

			test.WriteFile(t, s.basedir, file, content)
		default:
			t.Fatalf("unknown tree identifier: %d", spec[0])
		}
	}
}

// Git returns a git wrapper that is useful to run git commands safely inside
// the test env repo.
func (s S) Git() Git {
	return s.git
}

// BaseDir returns the base dir of the test env. All dirs/files created through
// the test env will be included inside this dir.
//
// It is a programming error to delete this dir, it will be automatically
// removed when the test finishes.
func (s S) BaseDir() string {
	return s.basedir
}

// CreateModule will create a module dir with the given relative path, returning
// a directory entry that can be used to create files inside the module dir.
func (s S) CreateModule(relpath string) DirEntry {
	t := s.t
	t.Helper()

	if filepath.IsAbs(relpath) {
		t.Fatalf("CreateModule() needs a relative path but given %q", relpath)
	}

	return newDirEntry(s.t, s.basedir, relpath)
}

// CreateStack will create a stack dir with the given relative path, returning a
// stack entry that can be used to create files inside the stack dir.
func (s S) CreateStack(relpath string) *StackEntry {
	t := s.t
	t.Helper()

	if filepath.IsAbs(relpath) {
		t.Fatalf("CreateStack() needs a relative path but given %q", relpath)
	}

	stack := &StackEntry{
		DirEntry: newDirEntry(t, s.basedir, relpath),
	}

	assert.NoError(t, terrastack.Init(stack.Path(), false))
	return stack
}

// CreateFile will create a file inside this dir entry with the given name and
// the given body. The body can be plain text or a format string identical to
// what is defined on Go fmt package.
//
// It returns a file entry that can be used to further manipulate the created
// file, like replacing its contents. The file entry is optimized for always
// replacing the file contents, not streaming (using file as io.Writer).
//
// If the file already exists its contents will be truncated, like os.Create
// behavior: https://pkg.go.dev/os#Create
func (de DirEntry) CreateFile(name, body string, args ...interface{}) *FileEntry {
	de.t.Helper()

	fe := &FileEntry{
		t:    de.t,
		path: filepath.Join(de.abspath, name),
	}
	fe.Write(body, args...)

	return fe
}

// Path returns the absolute path of the directory entry.
func (de DirEntry) Path() string {
	return de.abspath
}

// RelPath returns the relative path of the directory entry.
func (de DirEntry) RelPath() string {
	return de.relpath
}

// Write writes the given text body on the file, replacing its contents.
// The body can be plain text or a format string identical to what is defined on
// Go fmt package.
//
// It behaves like os.WriteFile: https://pkg.go.dev/os#WriteFile
func (fe FileEntry) Write(body string, args ...interface{}) {
	fe.t.Helper()

	body = fmt.Sprintf(body, args...)

	if err := os.WriteFile(fe.path, []byte(body), 0700); err != nil {
		fe.t.Fatalf("os.WriteFile(%q) = %v", fe.path, err)
	}
}

// Path returns the absolute path of the file.
func (fe FileEntry) Path() string {
	return fe.path
}

// ModSource returns the relative import path for the module with the given
// module dir entry. The path is relative to stack dir itself (hence suitable to
// be a module source path).
func (se StackEntry) ModSource(mod DirEntry) string {
	relpath, err := filepath.Rel(se.abspath, mod.abspath)
	assert.NoError(se.t, err)
	return relpath
}

// Path returns the absolute path of the stack.
func (se StackEntry) Path() string {
	return se.DirEntry.abspath
}

// RelPath returns the relative path of the stack. It is relative to the base
// dir of the test environment that created this stack.
func (se StackEntry) RelPath() string {
	return se.DirEntry.relpath
}

func newDirEntry(t *testing.T, basedir string, relpath string) DirEntry {
	t.Helper()

	abspath := filepath.Join(basedir, relpath)
	test.MkdirAll(t, abspath)

	return DirEntry{
		t:       t,
		abspath: abspath,
		relpath: relpath,
	}
}
