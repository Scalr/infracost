package schema

import (
	// nolint:gosec

	"fmt"
	"regexp"
)

type ProjectMetadata struct {
	Path               string `json:"path"`
	Type               string `json:"type"`
	VCSRepoURL         string `json:"vcsRepoUrl,omitempty"`
	VCSSubPath         string `json:"vcsSubPath,omitempty"`
	VCSPullRequestURL  string `json:"vcsPullRequestUrl,omitempty"`
	TerraformWorkspace string `json:"terraformWorkspace,omitempty"`
}

// Project contains the existing, planned state of
// resources and the diff between them.
type Project struct {
	Name          string
	Metadata      *ProjectMetadata
	PastResources []*Resource
	Resources     []*Resource
	Diff          []*Resource
	HasDiff       bool
}

func NewProject(name string, metadata *ProjectMetadata) *Project {
	return &Project{
		Name:     name,
		Metadata: metadata,
		HasDiff:  true,
	}
}

// AllResources returns a pointer list of all resources of the state.
func (p *Project) AllResources() []*Resource {
	var resources []*Resource
	resources = append(resources, p.PastResources...)
	resources = append(resources, p.Resources...)
	return resources
}

// CalculateDiff calculates the diff of past and current resources
func (p *Project) CalculateDiff() {
	if p.HasDiff {
		p.Diff = calculateDiff(p.PastResources, p.Resources)
	}
}

// AllProjectResources returns the resources for all projects
func AllProjectResources(projects []*Project) []*Resource {
	resources := make([]*Resource, 0)

	for _, p := range projects {
		resources = append(resources, p.Resources...)
	}

	return resources
}

func GenerateProjectName(metadata *ProjectMetadata) string {
	var n string

	// If the VCS repo is set, create the name from that
	if metadata.VCSRepoURL != "" {
		n = nameFromRepoURL(metadata.VCSRepoURL)

		if metadata.VCSSubPath != "" {
			n += "/" + metadata.VCSSubPath
		}
		// If not then use the filepath to the project
	} else {
		n = metadata.Path
	}

	if metadata.TerraformWorkspace != "" && metadata.TerraformWorkspace != "default" {
		n += fmt.Sprintf(" (%s)", metadata.TerraformWorkspace)
	}

	return n
}

// Parses the "org/repo" from the git URL if possible.
// Otherwise it just returns the URL.
func nameFromRepoURL(url string) string {
	r := regexp.MustCompile(`(?:\w+@|http(?:s)?:\/\/)(?:.*@)?([^:\/]+)[:\/]([^\.]+)`)
	m := r.FindStringSubmatch(url)

	if len(m) > 2 {
		var n = m[2]

		if m[1] == "dev.azure.com" || m[1] == "ssh.dev.azure.com" {
			n = parseAzureDevOpsRepoPath(m[2])
		}

		return n
	}

	return url
}

func parseAzureDevOpsRepoPath(path string) string {
	r := regexp.MustCompile(`(?:(.+)(?:\/base\/_git\/)(.+))`)
	m := r.FindStringSubmatch(path)

	if len(m) > 2 {
		return m[1] + "/" + m[2]
	}

	r = regexp.MustCompile(`(?:(?:v3\/)(.+)(?:\/base\/)(.+))`)
	m = r.FindStringSubmatch(path)

	if len(m) > 2 {
		return m[1] + "/" + m[2]
	}

	return path
}
