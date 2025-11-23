package codegov

import "time"

// GitHubRepository represents a GitHub repository from the API
type GitHubRepository struct {
	Name              string    `json:"name"`
	Description       string    `json:"description"`
	HTMLURL           string    `json:"html_url"`
	Private           bool      `json:"private"`
	Fork              bool      `json:"fork"`
	Archived          bool      `json:"archived"`
	Homepage          string    `json:"homepage"`
	Topics            []string  `json:"topics"`
	DefaultBranch     string    `json:"default_branch"`
	LanguagesURL      string    `json:"languages_url"`
	ReleasesURL       string    `json:"releases_url"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
	PushedAt          time.Time `json:"pushed_at"`
}

// GitHubLicense represents license information from GitHub API
type GitHubLicense struct {
	HTMLURL string `json:"html_url"`
	License struct {
		SPDXID string `json:"spdx_id"`
	} `json:"license"`
	Message string `json:"message"`
}

// GitHubRelease represents a release from GitHub API
type GitHubRelease struct {
	Prerelease bool   `json:"prerelease"`
	ZipballURL string `json:"zipball_url"`
	PublishedAt time.Time `json:"published_at"`
}

// License represents a license in code.gov format
type License struct {
	URL  string `json:"URL"`
	Name string `json:"name"`
}

// Contact represents contact information
type Contact struct {
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
	URL   string `json:"URL,omitempty"`
	Phone string `json:"phone,omitempty"`
}

// Permissions represents release permissions
type Permissions struct {
	Licenses  []License `json:"licenses"`
	UsageType string    `json:"usageType"`
}

// DateInfo represents date information for a release
type DateInfo struct {
	Created              string `json:"created"`
	LastModified         string `json:"lastModified"`
	MetadataLastUpdated  string `json:"metadataLastUpdated"`
}

// Release represents a single release in code.gov format
type Release struct {
	Name           string      `json:"name"`
	RepositoryURL  string      `json:"repositoryURL"`
	Description    string      `json:"description"`
	Permissions    Permissions `json:"permissions"`
	LaborHours     float64     `json:"laborHours"`
	Tags           []string    `json:"tags"`
	Contact        Contact     `json:"contact"`
	Status         string      `json:"status"`
	VCS            string      `json:"vcs"`
	HomepageURL    string      `json:"homepageURL"`
	DownloadURL    string      `json:"downloadURL"`
	DisclaimerURL  string      `json:"disclaimerURL,omitempty"`
	Languages      []string    `json:"languages,omitempty"`
	Date           DateInfo    `json:"date"`
}

// MeasurementType represents measurement type for code.gov
type MeasurementType struct {
	Method string `json:"method"`
}

// CodeGovJSON represents the complete code.gov JSON structure
type CodeGovJSON struct {
	Version         string          `json:"version"`
	Agency          string          `json:"agency"`
	MeasurementType MeasurementType `json:"measurementType"`
	Releases        []Release       `json:"releases"`
}

// OverrideAction represents an override action
type OverrideAction struct {
	Project  string `json:"project"`
	Action   string `json:"action"`
	Property string `json:"property,omitempty"`
	Value    interface{} `json:"value,omitempty"`
}

// OverrideJSON represents override configuration
type OverrideJSON struct {
	Overrides []OverrideAction `json:"overrides"`
}
