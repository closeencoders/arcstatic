package config

type SiteContext struct {
	ComponentMap map[string][]byte
	TemplateMap  map[string][]byte

	SiteUrl  string `yaml:"site_url"`
	SiteRoot string
	Base     string

	DefaultType string `yaml:"default_type"`

	SiteInputRoot string

	FrontmatterToken []byte `yaml:"frontmatter_token"`

	// Full context path to were post content files are loaded from. This can be outside of the root of the project
	PostInputDir string `yaml:"post_input_dir"`
	// The relative path to the rendered output for posts and misc content. This current is limited to the subdir inside of the root of the project
	PostOutDir   string `yaml:"post_output_dir"`
	PageInputDir string

	FullHtmlPath         bool
	GeneratePostMetadata bool
	GenerateSitemapXml   bool
	MakeTableOfContents  bool `yaml:"make_toc"`
}
