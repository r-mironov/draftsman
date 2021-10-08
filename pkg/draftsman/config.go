package draftsman

import "net/url"

type Config struct {
	TmpDir       string `mapstructure:"TMP_DIR"`
	GitlabHost   string `mapstructure:"GITLAB_HOST"`
	Token        string `mapstructure:"GITLAB_TOKEN"`
	ProjectPath  string `mapstructure:"PROJECT_PATH"`
	Ref          string `mapstructure:"REF"`
	GitlabCiFile string `mapstructure:"GITLAB_CI_FILE"`
}

// Config struct for gitlab-ci:include
type GitlabCiConfig struct {
	Include []IncludeLocation `yaml:"include"`
}

type IncludeLocation struct {
	Project string      `yaml:"project,omitempty"`
	Ref     string      `yaml:"ref,omitempty"`
	File    StringArray `yaml:"file,omitempty"`
	Remote  string      `yaml:"remote,omitempty"`
	Local   string      `yaml:"local,omitempty"`
}

type StringArray []string

type IncludeElement struct {
	Project string
	Id      string
	Ref     string
	File    string
}

type Node struct {
	Me    string
	Child string
}

type Graph []Node

var List = make(Graph, 0)

var AppConfig = Config{
	TmpDir: "tmd_dir",
}

var ProjectsListUrl = url.URL{
	Scheme:   "https",
	Path:     "/api/v4/projects",
	RawQuery: "search_namespaces=true&search=%v",
}

var ConfigUrl = url.URL{
	Scheme:   "https",
	Path:     "/api/v4/projects/%v/repository/files/%v/raw",
	RawPath:  "/api/v4/projects/%v/repository/files/%v/raw",
	RawQuery: "ref=",
}
