package main

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/closeencoders/arcstatic/internal/testutil"
	"github.com/closeencoders/arcstatic/source"
	"github.com/spf13/cobra"
)

func TestBuildFlags(t *testing.T) {

	siteRoot := "testroot"

	configLoc := filepath.Join(siteRoot, source.ConfigName)
	testPostLoc := filepath.Join(siteRoot, "_posts", "2026-06-01-post.md")

	siteOutputRoot := filepath.Join(siteRoot, "output")
	expectedPostFile := "testroot/output/post/index.html"

	ignoredConfig := []byte(fmt.Sprintf("site_output_root: %s", "ignore/this"))
	usedConfig := []byte(fmt.Sprintf("site_output_root: %s", siteOutputRoot))

	tests := []struct {
		name           string
		args           []string
		files          fstest.MapFS
		wantCliOutPath string
	}{
		{
			name: "Override Config With Output Dir Command With given Configuration",
			args: []string{"-i", siteRoot, "-o", siteOutputRoot, "-b"},
			files: fstest.MapFS{
				testPostLoc: &fstest.MapFile{Data: []byte("---\ntitle: test post a\n---\n# Header")},
				configLoc:   &fstest.MapFile{Data: ignoredConfig},
			},
			wantCliOutPath: siteOutputRoot,
		},
		{
			name: "Override Config With Output Dir Command With Default Configuration",
			args: []string{"-i", siteRoot, "-o", siteOutputRoot, "-b"},
			files: fstest.MapFS{
				testPostLoc: &fstest.MapFile{Data: []byte("---\ntitle: test post b\n---\n# Header")},
				// config file is not provided. Should Default.
			},
			wantCliOutPath: siteOutputRoot,
		},
		{
			name: "Use Config Output Dir When Output Path Is Empty",
			args: []string{"-i", siteRoot, "-o", "", "-b"},
			files: fstest.MapFS{
				testPostLoc: &fstest.MapFile{Data: []byte("---\ntitle: test post c\n---\n# Header")},
				configLoc:   &fstest.MapFile{Data: usedConfig},
			},
			wantCliOutPath: "",
		},
		{
			name: "Use Config Output Dir When Output Path Is Blank",
			args: []string{"-i", siteRoot, "-o", "       ", "-b"},
			files: fstest.MapFS{
				testPostLoc: &fstest.MapFile{Data: []byte("---\ntitle: test post c\n---\n# Header")},
				configLoc:   &fstest.MapFile{Data: usedConfig},
			},
			wantCliOutPath: "",
		},
		{
			name: "Use Config Output Dir When Output Path Is Not Provided",
			args: []string{"-i", siteRoot, "-b"},
			files: fstest.MapFS{
				testPostLoc: &fstest.MapFile{Data: []byte("---\ntitle: test post d\n---\n# Header")},
				configLoc:   &fstest.MapFile{Data: usedConfig},
			},
			wantCliOutPath: "",
		},
		{
			name: "Use Config Output Dir When No Paths Provided And Input Defaults to current working dir",
			args: []string{"-b"},
			files: fstest.MapFS{
				testPostLoc: &fstest.MapFile{Data: []byte("---\ntitle: test post e\n---\n# Header")},
				configLoc:   &fstest.MapFile{Data: usedConfig},
			},
			wantCliOutPath: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			store := testutil.NewFakeStorageWithState(
				test.files,
				&testutil.FakeState{CurrentWd: siteRoot, WrittenFiles: []string{}},
			)

			rootCmd := &cobra.Command{}
			ssg := NewSsg(rootCmd, store)
			ssg.cmd.SetArgs(test.args)

			buf := new(bytes.Buffer)
			ssg.cmd.SetErr(buf)
			ssg.cmd.SetOut(buf)

			if err := ssg.cmd.Execute(); err != nil {
				t.Fatalf("invalid test commands: %v", err)
			}

			testutil.AssertEqual(t, "CLI outputPath", ssg.outputLocation, test.wantCliOutPath)
			testutil.AssertEqual(t, "Context outputPath", ssg.ctx.SiteOutputRoot, siteOutputRoot)
			if len(store.State().WrittenFiles) < 1 {
				t.Fatal("no files written in test")
			}
			testutil.AssertEqual(t, "Expected Post File", store.State().WrittenFiles[0], expectedPostFile)
		})
	}
}

func TestBuildReachesGenerator(t *testing.T) {

	tests := []struct {
		name          string
		fileName      string
		fileData      []byte
		count         int
		allowAltFiles bool
	}{
		{
			name:  "Should Not Attempt To Write Anything Without Files",
			count: 0,
		},
		{
			name: "Should Attempt To Write Basic Content And Alt Files",

			fileName: "fakepostloc/2026-06-01-basic_post.md",
			fileData: []byte("---\ntitle: Hello\n---\ntest"),

			allowAltFiles: true,

			count: 3,
		},
		{
			name: "Should Attempt To Write Basic Content",

			fileName: "fakepostloc/2026-06-01-basic_post.md",
			fileData: []byte("---\ntitle: Hello\n---\ntest"),

			count: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			testRoot := "rootpath"
			ctx := source.CreateDefaultContext(testRoot)
			ctx.PostInputDir = "fakepostloc"
			ctx.MakePostMetadata = test.allowAltFiles
			ctx.MakeSitemapXML = test.allowAltFiles

			fileMap := fstest.MapFS{test.fileName: &fstest.MapFile{Data: test.fileData}}
			store, _ := testutil.NewFakeStorage(fileMap)

			if err := executeBuild(ctx, store); err != nil {
				t.Fatalf("unexpected error while testing build output %v", err)
			}

			if len(store.State().WrittenFiles) != test.count {
				t.Errorf("the correct number of files was not written to fake storage for test wnt: %d got: %d", test.count, len(store.State().WrittenFiles))
			}
			if ctx.SiteRoot != testRoot {
				t.Errorf("context did not pickup the correct root location: wnt: %s got: %s", testRoot, ctx.SiteRoot)
			}
			if len(store.State().WrittenFiles) == 1 {
				name := store.State().WrittenFiles[0]
				if !strings.HasPrefix(name, ctx.SiteRoot) {
					t.Errorf("config root location was not applied to content path: wnt: %s got %s", ctx.SiteRoot, name)
				}
			}
		})
	}
}

// TODO: sort restriction on prefix
func TestWithDatePrefixWriteOrder(t *testing.T) {

	files := fstest.MapFS{
		"fakepostloc/1995-01-01-8-post.md": &fstest.MapFile{Data: []byte("---\ntitle: 8\ndate: 1995-01-01\n---\n# Header")},
		"fakepostloc/2020-06-10-7-post.md": &fstest.MapFile{Data: []byte("---\ntitle: 7\ndate: 2020-06-10\n---\n# Header")},
		"fakepostloc/2020-09-10-6-post.md": &fstest.MapFile{Data: []byte("---\ntitle: 6\ndate: 2020-09-10\n---\n# Header")},
		"fakepostloc/2020-10-10-5-post.md": &fstest.MapFile{Data: []byte("---\ntitle: 5\ndate: 2020-10-10\n---\n# Header")},
		"fakepostloc/2021-11-10-4-post.md": &fstest.MapFile{Data: []byte("---\ntitle: 4\ndate: 2021-11-10\n---\n# Header")},
		"fakepostloc/2022-10-11-3-post.md": &fstest.MapFile{Data: []byte("---\ntitle: 3\ndate: 2022-10-11\n---\n# Header")},
		"fakepostloc/2023-10-10-2-post.md": &fstest.MapFile{Data: []byte("---\ntitle: 2\ndate: 2023-10-10\n---\n# Header")},
		"fakepostloc/2024-10-10-1-post.md": &fstest.MapFile{Data: []byte("---\ntitle: 1\ndate: 2024-10-10\n---\n# Header")},
	}

	expected := []string{
		"rootpath/8-post/index.html",
		"rootpath/7-post/index.html",
		"rootpath/6-post/index.html",
		"rootpath/5-post/index.html",
		"rootpath/4-post/index.html",
		"rootpath/3-post/index.html",
		"rootpath/2-post/index.html",
		"rootpath/1-post/index.html",
	}

	ctx := source.CreateDefaultContext("rootpath")
	ctx.PostInputDir = "fakepostloc"
	ctx.MakePostMetadata = false
	ctx.MakeSitemapXML = false

	store, _ := testutil.NewFakeStorage(files)
	err := executeBuild(ctx, store)
	if err != nil {
		t.Fatalf("build test failed with: %v", err)
	}

	if len(expected) != len(store.State().WrittenFiles) {
		t.Fatalf("the correct number of files was not written to fake storage for test wnt: %d got: %d", len(expected), len(store.State().WrittenFiles))
	}

	for i, name := range expected {
		if store.State().WrittenFiles[i] != name {
			t.Fatalf("Wrong order wnt: %s, got: %s", name, store.State().WrittenFiles[i])
		}
	}
}

func TestNoDatePrefixWriteOrder(t *testing.T) {

	files := fstest.MapFS{
		"fakepostloc/8-post.md": &fstest.MapFile{Data: []byte("---\ntitle: 8\ndate: 1995-01-01\n---\n# Header")},
		"fakepostloc/2-post.md": &fstest.MapFile{Data: []byte("---\ntitle: 2\ndate: 2023-10-10\n---\n# Header")},
		"fakepostloc/5-post.md": &fstest.MapFile{Data: []byte("---\ntitle: 5\ndate: 2020-10-10\n---\n# Header")},
		"fakepostloc/3-post.md": &fstest.MapFile{Data: []byte("---\ntitle: 3\ndate: 2022-10-11\n---\n# Header")},
		"fakepostloc/4-post.md": &fstest.MapFile{Data: []byte("---\ntitle: 4\ndate: 2021-11-10\n---\n# Header")},
		"fakepostloc/1-post.md": &fstest.MapFile{Data: []byte("---\ntitle: 1\ndate: 2024-10-10\n---\n# Header")},
		"fakepostloc/7-post.md": &fstest.MapFile{Data: []byte("---\ntitle: 7\ndate: 2020-06-10\n---\n# Header")},
		"fakepostloc/6-post.md": &fstest.MapFile{Data: []byte("---\ntitle: 6\ndate: 2020-09-10\n---\n# Header")},
	}

	expected := []string{
		"rootpath/1-post/index.html",
		"rootpath/2-post/index.html",
		"rootpath/3-post/index.html",
		"rootpath/4-post/index.html",
		"rootpath/5-post/index.html",
		"rootpath/6-post/index.html",
		"rootpath/7-post/index.html",
		"rootpath/8-post/index.html",
	}

	ctx := source.CreateDefaultContext("rootpath")
	ctx.PostInputDir = "fakepostloc"
	ctx.MakePostMetadata = false
	ctx.MakeSitemapXML = false

	// invokes a sort function mid exec by the provided frontmatter date instead of the file name.
	ctx.AllowNamelessDateSort = true

	store, _ := testutil.NewFakeStorage(files)
	err := executeBuild(ctx, store)
	if err != nil {
		t.Fatalf("build test failed with: %v", err)
	}

	if len(expected) != len(store.State().WrittenFiles) {
		t.Fatalf("the correct number of files was not written to fake storage for test wnt: %d got: %d", len(expected), len(store.State().WrittenFiles))
	}

	for i, name := range expected {
		if store.State().WrittenFiles[i] != name {
			t.Fatalf("Wrong order wnt: %s, got: %s", name, store.State().WrittenFiles[i])
		}
	}
}
