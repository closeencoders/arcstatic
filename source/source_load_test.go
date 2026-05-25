package source

import (
	"io/fs"
	"path"
	"sync"
	"testing"
	"testing/fstest"

	"github.com/closeencoders/arcstatic/config"
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

	t.Parallel()

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

	var wg sync.WaitGroup

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			wg.Add(1)

			mockStore := fakeStorage{testFiles: tt.configFile}
			result, err := LoadSiteContext(baseDir, mockStore)

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

func TestLoadMetadata(t *testing.T) {

	t.Parallel()

	tests := []struct {
		name         string
		fileData     fstest.MapFS
		path         string
		wantTitle    string
		wantPath     string
		wantTemplate string
		notLoaded    bool
	}{
		{
			name: "Known Post Location Should Load",
			path: "fakepostloc",
			fileData: fstest.MapFS{
				"fakepostloc/basic_post.md": &fstest.MapFile{Data: []byte("---\ntitle: Hello\n---\n# Header")},
			},
			wantTitle:    "Hello",
			wantPath:     "basic-post",
			wantTemplate: "post.html",
		},
		{
			name: "Known Post Location Should Load With Irregular Name",
			path: "fakepostloc",
			fileData: fstest.MapFS{
				"fakepostloc/basic *&^(*&)  post.md": &fstest.MapFile{Data: []byte("---\ntitle: Hello\n---\n# Header")},
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
		{
			name: "Unknown File Extension Should Not Load Or Attempt To Load Anything",
			path: "fakepageloc",
			fileData: fstest.MapFS{
				"fakepostloc/about.xyz": &fstest.MapFile{Data: []byte("---\ntitle: About\n---\n# Header")},
			},
			notLoaded: true,
		},
		{
			name: "Unknown Location Should Not Load Or Attempt To Load Anything",
			path: "fakepageloc",
			fileData: fstest.MapFS{
				"xxxxxx/about.html": &fstest.MapFile{Data: []byte("---\ntitle: About\n---\n# Header")},
			},
			notLoaded: true,
		},
		{
			name: "Draft Post Should Not Load When Set To True",
			path: "fakepostloc",
			fileData: fstest.MapFS{
				"fakepostloc/draft_post.md": &fstest.MapFile{Data: []byte("---\ntitle: Hello\ndraft: true\n---\n# Header")},
			},
			notLoaded: true,
		},
	}

	var wg sync.WaitGroup

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			wg.Add(1)

			ctx := createDefaultContext("public")
			ctx.PostInputDir = "fakepostloc"
			ctx.PageInputDir = "fakepageloc"

			mockStore := fakeStorage{testFiles: tt.fileData}
			ml := metadata{ctx: ctx, store: mockStore}
			result, err := ml.LoadMetadata(tt.path)
			if err != nil {
				t.Fatalf("LoadMetadata() execution unexpected error = %v", err)
			}

			entitiesCount := len(result.SiteContentEntities)
			if tt.notLoaded {
				if entitiesCount != 0 {
					t.Fatalf("expected 0 entities loaded, got %d", entitiesCount)
				}
				return
			}
			if entitiesCount == 0 {
				t.Fatalf("metadata should have been loaded for target path: %q", tt.path)
			}

			ce := result.SiteContentEntities[0]
			AssertEqual(t, "template", ce.ContentMetadata.TemplateId, tt.wantTemplate)
			AssertEqual(t, "title", ce.ContentMetadata.Title, tt.wantTitle)
		})
	}
}

func TestSiteManifest(t *testing.T) {

	t.Parallel()

	tests := []struct {
		name                string
		Ce                  ContentEntity
		wantTypes           []string
		DefaultTypeOverride string
	}{
		{
			name: "No Type Is Set, Should Still Have Default Collection",
			Ce: ContentEntity{
				Name:            "test",
				ContentMetadata: ContentMetadata{},
			},
			wantTypes: []string{"Posts"},
		},
		{
			name: "Type Is Set, Should Have Default Collection And Defined Type",
			Ce: ContentEntity{
				Name:            "test",
				ContentMetadata: ContentMetadata{Type: "Blogs"},
			},
			wantTypes: []string{"Blogs", "Posts"},
		},
		{
			name: "Overridden Default Content Type Should Be Returned",
			Ce: ContentEntity{
				Name: "test",
				ContentMetadata: ContentMetadata{
					Type: "Blogs",
				},
			},
			DefaultTypeOverride: "Other",
			wantTypes:           []string{"Blogs", "Other"},
		},
	}

	var wg sync.WaitGroup

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			wg.Add(1)

			ctx := config.NewContext("testlocation")
			if tt.DefaultTypeOverride != "" {
				ctx.DefaultType = tt.DefaultTypeOverride
			}

			manifest := NewManifest(*ctx, []*ContentEntity{&tt.Ce})

			if len(tt.wantTypes) > 0 {
				for _, expectedType := range tt.wantTypes {
					_, exists := manifest[expectedType]
					if !exists {
						t.Errorf("Expected manifest to have type of: %s", expectedType)
					}
				}
			}
		})
	}
}

func AssertEqual(t *testing.T, msg string, got, want any) {
	t.Helper()
	if got != want {
		t.Errorf("%s:\nGot:[%v]\nWnt:[%v]", msg, got, want)
	}
}
