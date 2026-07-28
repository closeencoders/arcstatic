package testutil

import (
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/closeencoders/arcstatic/storage"
)

type FakeStorage struct {
	TestFiles fstest.MapFS
	state     *FakeState
}

// TODO:
type FakeState struct {
	WrittenFiles []string
}

func NewFakeStorage(testFiles fstest.MapFS) FakeStorage {
	return FakeStorage{testFiles, &FakeState{WrittenFiles: []string{}}}
}

func NewFakeStorageWithState(testFiles fstest.MapFS, state *FakeState) FakeStorage {
	return FakeStorage{testFiles, state}
}

var _ storage.Storage = FakeStorage{}

func (fs FakeStorage) State() FakeState {
	return *fs.state
}

func (fs FakeStorage) Write(name string, data []byte, perm int) error {
	fs.state.WrittenFiles = append(fs.state.WrittenFiles, name)
	return nil
}

func (fs FakeStorage) Open(name string) (fs.File, error) {
	return fs.TestFiles.Open(name)
}

func (fs FakeStorage) Mkdir(perm int, path ...string) error {
	return nil
}

func (fs FakeStorage) CopyDir(perm int, from string, to string) error {
	return nil
}

func (fs FakeStorage) Copy(perm int, from string, to string) error {
	return nil
}

func AssertEqual(t *testing.T, msg string, got, want any) {
	t.Helper()
	if got != want {
		t.Errorf("%s:\nGot:[%v]\nWnt:[%v]", msg, got, want)
	}
}
