package source

import (
	"io/fs"
	"path"
	"testing"
	"testing/fstest"

	"github.com/closeencoders/arcstatic/storage"
)

type fakeStorage struct {
	testFiles fstest.MapFS
}

var _ storage.Storage = fakeStorage{}

func (fs fakeStorage) Write(name string, data []byte, perm int) error {
	return nil
}

func (fs fakeStorage) Open(name string) (fs.File, error) {
	return fs.testFiles.Open(name)
}

func (fs fakeStorage) Mkdir(perm int, path ...string) (string, error) {
	return "", nil
}

func TestLoadConfig(t *testing.T) {

	const baseDir = "testlocation"
	defaultPostLoc := path.Join(baseDir, _postsLoc)
	configLoc := path.Join(baseDir, _configName)

	tests := []struct {
		name       string
		configFile fstest.MapFS
		wantUrl    string
		wantDir    string
		wantErr    bool
	}{
		{
			name: "Context Should Contain Partial Override Values From File",
			configFile: fstest.MapFS{
				configLoc: &fstest.MapFile{
					Data: []byte("site_url: https://location.com\npost_input_dir: location123/posts"),
				},
			},
			wantUrl: "https://location.com",
			wantDir: "location123/posts",
		},
		{
			name: "Invalid Configuration Should Return Error",
			configFile: fstest.MapFS{
				configLoc: &fstest.MapFile{Data: []byte("site_url: : : : :")},
			},
			wantErr: true,
		},
		{
			name: "Empty Config Should Use Full Default",
			configFile: fstest.MapFS{
				configLoc: &fstest.MapFile{Data: []byte("")},
			},
			wantUrl: _defaultUrl,
			wantDir: defaultPostLoc,
		},
		{
			name:       "File Not Found Should Use Full Default",
			configFile: nil,
			wantUrl:    _defaultUrl,
			wantDir:    defaultPostLoc,
		},
	}

	for _, tt := range tests {
		tc := tt

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mockStore := fakeStorage{testFiles: tc.configFile}
			result, err := LoadSiteContext(baseDir, mockStore)

			if err == nil && tc.wantErr {
				t.Fatalf("LoadSiteContext() unexpected error state:\nGot: [%v]\nWanted: [%v]", err, tc.wantErr)
			}
			if tc.wantErr && err != nil {
				return
			}
			AssertEqual(t, "SiteUrl", result.SiteURL, tc.wantUrl)
			AssertEqual(t, "PostInputDir", result.PostInputDir, tc.wantDir)
		})
	}
}

func TestLoadMetadata(t *testing.T) {

	ctx := createDefaultContext("public")
	ctx.SiteRoot = "public"
	ctx.PostInputDir = "fakepostloc"
	ctx.PageInputDir = "fakepageloc"

	tests := []struct {
		name         string
		fileData     fstest.MapFS
		path         string
		wantTitle    string
		wantPath     string
		wantTemplate string
		// Data is not valid or configured to not be loaded as content, but the process should continue
		notLoaded bool
	}{
		{
			name: "Known Post Location Should Load",
			path: ctx.PostInputDir,
			fileData: fstest.MapFS{
				"fakepostloc/basic_post.md": &fstest.MapFile{Data: []byte("---\ntitle: Hello\n---\n# Header")},
			},
			wantTitle:    "Hello",
			wantPath:     "basic-post",
			wantTemplate: "post.html",
		},
		{
			name: "Known Post Location Should Load With Irregular Name",
			path: ctx.PostInputDir,
			fileData: fstest.MapFS{
				"fakepostloc/basic *&^(*&)  post.md": &fstest.MapFile{Data: []byte("---\ntitle: Hello\n---\n# Header")},
			},
			wantTitle:    "Hello",
			wantPath:     "basic-post",
			wantTemplate: "post.html",
		},
		{
			name: "Known Page Location Should Load",
			path: ctx.PageInputDir,
			fileData: fstest.MapFS{
				"fakepageloc/about.html": &fstest.MapFile{Data: []byte("---\ntitle: About\n---\n# Header")},
			},
			wantTitle:    "About",
			wantPath:     "about",
			wantTemplate: "page.html",
		},
		{
			name: "Unknown Location Should Not Load Or Attempt To Load Anything",
			path: ctx.PageInputDir,
			fileData: fstest.MapFS{
				"xxxxxx/about.html": &fstest.MapFile{Data: []byte("---\ntitle: About\n---\n# Header")},
			},
			notLoaded: true,
		},
		{
			name: "Draft Post Should Not Load When Set To True",
			path: ctx.PostInputDir,
			fileData: fstest.MapFS{
				"fakepostloc/draft_post.md": &fstest.MapFile{Data: []byte("---\ntitle: Hello\ndraft: true\n---\n# Header")},
			},
			notLoaded: true,
		},
	}

	for _, tt := range tests {
		tc := tt

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mockStore := fakeStorage{testFiles: tc.fileData}
			ml := metadata{ctx: ctx, store: mockStore}
			result, err := ml.LoadMetadata(tc.path)
			if err != nil {
				t.Fatalf("LoadMetadata() execution unexpected error = %v", err)
			}

			entitiesCount := len(result.SiteContentEntities)
			if tc.notLoaded {
				if entitiesCount != 0 {
					t.Fatalf("expected 0 entities loaded, got %d", entitiesCount)
				}
				// Gracefully exit execution stack, validation complete
				return
			}
			if entitiesCount == 0 {
				t.Fatalf("invariant violation: metadata should have been loaded for target path: %q", tc.path)
			}

			ce := result.SiteContentEntities[0]
			AssertEqual(t, "template", ce.ContentMetadata.TemplateId, tc.wantTemplate)
			AssertEqual(t, "title", ce.ContentMetadata.Title, tc.wantTitle)
		})
	}
}

func AssertEqual(t *testing.T, msg string, got, want any) {
	t.Helper()
	if got != want {
		t.Errorf("%s:\nGot:[%v]\nWnt:[%v]", msg, got, want)
	}
}
