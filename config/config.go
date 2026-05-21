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

	// FrontmatterToken marks the boundary of configuration blocks in source files (e.g., "---" or "+++").
	FrontmatterToken string `yaml:"frontmatter_token"`

	// PostInputDir is the absolute path to the raw posts folder, decoupled from output locations.
	PostInputDir string `yaml:"post_input_dir"`

	// PostOutputDir is the path to write compiled posts, relative to the SiteRoot.
	PostOutputDir string `yaml:"post_output_dir"`

	// TODO:
	PageInputDir string

	// TODO: Isolation to individual content preferences
	FullHtmlPath        bool
	MakePostMetadata    bool `yaml:"make_post_metadata"`
	MakeSitemapXML      bool `yaml:"make_sitemap"`
	MakeTableOfContents bool `yaml:"make_toc"`
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

		FullHtmlPath:        false,
		MakeSitemapXML:      true,
		MakePostMetadata:    true,
		MakeTableOfContents: false,
	}
}
