package source

import (
	"io/fs"
	"path"
	"testing"
	"testing/fstest"

	"github.com/closeencoders/arcstatic/storage"
)

type FakeStorage struct {
	TestFiles fstest.MapFS
}

var _ storage.Storage = FakeStorage{}

func (fs FakeStorage) Write(name string, data []byte, perm int) error {
	return nil
}

func (fs FakeStorage) Open(name string) (fs.File, error) {
	return fs.TestFiles.Open(name)
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
			name:       "Config File Not Found Should Use Full Default",
			configFile: nil,
			wantUrl:    _defaultUrl,
			wantDir:    defaultPostLoc,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			fs := FakeStorage{
				TestFiles: tt.configFile,
			}
			result, err := LoadSiteContext(baseDir, fs)

			if (err != nil) != tt.wantErr {
				t.Fatalf("unexpected error state while loading site context test: %v, wantErr: %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			AssertEqual(t, "SiteUrl", result.SiteURL, tt.wantUrl)
			AssertEqual(t, "PostInputDir", result.PostInputDir, tt.wantDir)
		})
	}
}

func TestTaxonomy(t *testing.T) {

	tests := []struct {
		name string

		path     string
		fileName string
		fileData []byte

		wantType string
	}{
		{
			name:     "Post With Default Type Should Be Added To Manifest",
			path:     "fakepostloc",
			fileName: "fakepostloc/2026-06-01-basic_post.md",
			fileData: []byte("---\ntitle: Hello\n---\n# Header"),

			wantType: "Posts",
		},
		{
			name:     "Post Valid Override Type Should Be Added To Manifest",
			path:     "fakepostloc",
			fileName: "fakepostloc/2026-06-01-basic_post.md",
			fileData: []byte("---\ntitle: Hello\ntype: other---\n# Header"),

			wantType: "other",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctx := createDefaultContext("public")
			ctx.PostInputDir = "fakepostloc"
			ctx.PageInputDir = "fakepageloc"

			testFile := fstest.MapFS{tt.fileName: &fstest.MapFile{Data: tt.fileData}}
			fs := FakeStorage{testFile}
			ml := NewMetadata(ctx, fs)

			result, err := ml.LoadMetadata(tt.path)
			if err != nil {
				t.Fatalf("LoadMetadata() execution unexpected error = %v", err)
			}
			if len(result.SiteContentEntities) == 0 {
				t.Fatalf("Expected Metadata for taxonomy test did not load")
			}

			ce := result.SiteContentEntities[0]
			AssertEqual(t, "wrong type", ce.ContentMetadata.Type, tt.wantType)

			_, ok := result.ContentManifest[tt.wantType]
			if !ok {
				t.Errorf("type was not applied to content manifest %s %s", tt.wantType, ce.FileName)
			}
		})
	}
}

// TODO: Refactor to break test up into chunks of expected behavior
func TestLoadMetadata(t *testing.T) {

	tests := []struct {
		name     string
		fileData fstest.MapFS
		path     string
	}{
		{
			name: "Unknown File Extension Should Not Load Or Attempt To Load Anything",
			path: "fakepageloc",
			fileData: fstest.MapFS{
				"fakepostloc/2026-06-01-about.xyz": &fstest.MapFile{Data: []byte("---\ntitle: About\n---\n# Header")},
			},
		},
		{
			name: "Draft Post Should Not Load When Set To True",
			path: "fakepostloc",
			fileData: fstest.MapFS{
				"fakepostloc/2026-06-01-draft_post.md": &fstest.MapFile{Data: []byte("---\ntitle: Hello\ndraft: true\n---\n# Header")},
			},
		},
		{
			name: "Unknown Page Location Should Not Load Or Attempt To Load Anything",
			path: "fakepageloc",
			fileData: fstest.MapFS{
				"xxxxxx/about.html": &fstest.MapFile{Data: []byte("---\ntitle: About\n---\n# Header")},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctx := createDefaultContext("public")
			ctx.PostInputDir = "fakepostloc"
			ctx.PageInputDir = "fakepageloc"

			fs := FakeStorage{tt.fileData}
			ml := metadata{ctx: ctx, store: fs}

			result, err := ml.LoadMetadata(tt.path)
			if err != nil {
				t.Fatalf("LoadMetadata() execution unexpected error = %v", err)
			}

			entitiesCount := len(result.SiteContentEntities)
			if entitiesCount != 0 {
				t.Errorf("expected 0 entities loaded, got %d", entitiesCount)
			}
		})
	}
}

func TestLoadValidMetadata(t *testing.T) {

	tests := []struct {
		name         string
		fileData     fstest.MapFS
		path         string
		wantTitle    string
		wantPath     string
		wantTemplate string
		wantType     string
	}{
		{
			name: "Known Post Location Should Load",
			path: "fakepostloc",
			fileData: fstest.MapFS{
				"fakepostloc/2026-06-01-basic_post.md": &fstest.MapFile{Data: []byte("---\ntitle: Hello\n---\n# Header")},
			},
			wantTitle:    "Hello",
			wantPath:     "basic-post",
			wantTemplate: "post.html",
		},
		{
			name: "Known Post Location Should Load With Irregular Name",
			path: "fakepostloc",
			fileData: fstest.MapFS{
				"fakepostloc/2026-06-01-basic *&^(*&)  post.md": &fstest.MapFile{Data: []byte("---\ntitle: Hello\n---\n# Header")},
			},
			wantTitle:    "Hello",
			wantPath:     "basic-post",
			wantTemplate: "post.html",
		},
		{
			name: "Known Page Location Should Load",
			path: "fakepageloc",
			fileData: fstest.MapFS{
				"fakepageloc/about.html": &fstest.MapFile{Data: []byte("---\ntitle: About\n---\n# Header")},
			},
			wantTitle:    "About",
			wantPath:     "about",
			wantTemplate: "page.html",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			ctx := createDefaultContext("public")
			ctx.PostInputDir = "fakepostloc"
			ctx.PageInputDir = "fakepageloc"

			fs := FakeStorage{test.fileData}
			ml := metadata{ctx: ctx, store: fs}

			result, err := ml.LoadMetadata(test.path)
			if err != nil {
				t.Fatalf("source loading unexpected error = %v", err)
			}

			if len(result.SiteContentEntities) == 0 {
				t.Fatalf("metadata should have been loaded for target path: %q", test.path)
			}

			ce := result.SiteContentEntities[0]
			if ce.InputPath == ctx.PostInputDir {
				if ce.ContentMetadata.Type == "" {
					AssertEqual(t, "type should have defaulted", ce.ContentMetadata.Type, ctx.DefaultType)
				} else {
					AssertEqual(t, "type", ce.ContentMetadata.Type, test.wantType)
				}
			}
			AssertEqual(t, "template", ce.ContentMetadata.TemplateId, test.wantTemplate)
			AssertEqual(t, "title", ce.ContentMetadata.Title, test.wantTitle)
		})
	}
}
