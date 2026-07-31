package testutil

import (
	"io/fs"
	"os"
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
	CurrentWd    string
	WrittenFiles []string
}

func NewFakeStorage(testFiles fstest.MapFS) (FakeStorage, error) {
	wd, err := os.Getwd()
	if err != nil {
		return FakeStorage{}, err
	}
	return FakeStorage{
		testFiles,
		&FakeState{
			CurrentWd:    wd,
			WrittenFiles: []string{},
		},
	}, nil
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

func (fs FakeStorage) GetWd() (string, error) {
	return fs.state.CurrentWd, nil
}

func AssertEqual(t *testing.T, msg string, got, want any) {
	t.Helper()
	if got != want {
		t.Errorf("%s:\nGot:[%v]\nWnt:[%v]", msg, got, want)
	}
}
