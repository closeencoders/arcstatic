package config

type SiteContext struct {
	// ComponentMap holds raw template partials or blocks keyed by their identifier name.
	ComponentMap map[string][]byte
	// TemplateMap holds raw layout templates keyed by their identifier name.
	TemplateMap map[string][]byte

	// SiteURL is the absolute base URL (e.g., "https://example.com").
	SiteURL  string `yaml:"site_url"`
	SiteRoot string
	Base     string

	DefaultType string `yaml:"default_type"`

	SiteOutputRoot string `yaml:"site_output_root"`

	// FrontmatterToken marks the boundary of configuration blocks in source files (e.g., "---" or "+++").
	FrontmatterToken string `yaml:"frontmatter_token"`
	// PostInputDir is the absolute path to the raw posts folder, decoupled from output locations.
	PostInputDir string `yaml:"post_input_dir"`
	// PostOutputDir is the path to write compiled posts, relative to the SiteRoot.
	PostOutputDir string `yaml:"post_output_dir"`
	// TODO:
	PageInputDir string

	// If set to true, will use a full path to the html file (e.g. /blah/content.html)
	FullHtmlPaths bool `yaml:"full_html_paths"`

	MaxDescriptionLen int `yaml:"max_description_len"`

	// TODO: Isolation to individual content preferences
	MakePostMetadata    bool `yaml:"make_post_metadata"`
	MakeSitemapXML      bool `yaml:"make_sitemap"`
	MakeTableOfContents bool `yaml:"make_toc"`

	JsonLog  bool   `yaml:"json_log"`
	LogLevel string `yaml:"log_level"`

	// By default, content is given in date order by file name to improve performance. This flag allows this to be disabled at the cost of performance.
	AllowNamelessDateSort bool `yaml:"allow_nameless_date_sort"`

	// TODO:
	KeepDateUrl bool `yaml:"keep_date_url"`
	// TODO:
	AllowTaxonomyPaths bool `yaml:"allow_taxonomy_paths"`

	AllowManifest bool
}

// TODO: enforce access patterns based on when data is mutated.
func NewContext(root string) *SiteContext {
	return &SiteContext{
		ComponentMap: make(map[string][]byte),
		TemplateMap:  make(map[string][]byte),

		SiteRoot: root,

		PostOutputDir: "/",
		Base:          "/",

		FrontmatterToken: "---",

		MaxDescriptionLen: 156,

		DefaultType: "All",

		MakeSitemapXML:   true,
		MakePostMetadata: true,

		AllowManifest: true,
	}
}
