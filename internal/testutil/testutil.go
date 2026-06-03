package testutil

import (
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/closeencoders/arcstatic/storage"
)

func NewFakeStorage(TestFiles fstest.MapFS) *FakeStorage {
	return &FakeStorage{testFiles: TestFiles}
}

type FakeStorage struct {
	testFiles fstest.MapFS
}

var _ storage.Storage = FakeStorage{}

func (fs FakeStorage) Write(name string, data []byte, perm int) error {
	return nil
}

func (fs FakeStorage) Open(name string) (fs.File, error) {
	return fs.testFiles.Open(name)
}

func (fs FakeStorage) Mkdir(perm int, path ...string) (string, error) {
	return "", nil
}

func AssertEqual(t *testing.T, msg string, got, want any) {
	t.Helper()
	if got != want {
		t.Errorf("%s:\nGot:[%v]\nWnt:[%v]", msg, got, want)
	}
}
