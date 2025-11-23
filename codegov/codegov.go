package codegov

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"
)

const (
	GitHubBaseURI = "https://api.github.com"
	OAuthTokenEnv = "OAUTH_TOKEN"
)

// SetOAuthToken sets the OAuth token in environment variable
func SetOAuthToken(token string) error {
	if !regexp.MustCompile(`^([0-9a-f]{40}){0,1}$`).MatchString(token) {
		return fmt.Errorf("invalid token format")
	}
	return os.Setenv(OAuthTokenEnv, token)
}

// GetOAuthToken retrieves the OAuth token from environment variable
func GetOAuthToken() string {
	token := os.Getenv(OAuthTokenEnv)
	return token
}

// TestOAuthToken validates the OAuth token
func TestOAuthToken(token ...string) bool {
	var tokenToTest string

	if len(token) > 0 {
		tokenToTest = token[0]
	} else {
		tokenToTest = GetOAuthToken()
	}

	if tokenToTest == "" {
		return false
	}

	return regexp.MustCompile(`^([0-9a-f]{40}){1}$`).MatchString(tokenToTest)
}

// TestURL verifies a URL is accessible
func TestURL(urlStr string) bool {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("HEAD", urlStr, nil)
	if err != nil {
		return false
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// GetGitHubRepositories fetches all repositories for an organization
func GetGitHubRepositories(organization string) ([]GitHubRepository, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	uri := fmt.Sprintf("%s/orgs/%s/repos?per_page=100", GitHubBaseURI, strings.ToLower(organization))

	var allRepos []GitHubRepository
	page := 1

	for {
		pageURL := fmt.Sprintf("%s&page=%d", uri, page)
		repos, hasNext, err := fetchRepositoriesPage(client, pageURL)
		if err != nil {
			return nil, err
		}

		allRepos = append(allRepos, repos...)

		if !hasNext {
			break
		}
		page++
	}

	return allRepos, nil
}

func fetchRepositoriesPage(client *http.Client, uri string) ([]GitHubRepository, bool, error) {
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, false, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/vnd.github.mercy-preview+json")

	if TestOAuthToken() {
		req.Header.Set("Authorization", fmt.Sprintf("token %s", GetOAuthToken()))
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, false, fmt.Errorf("request failed with status code %d: %s", resp.StatusCode, string(body))
	}

	var repos []GitHubRepository
	if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
		return nil, false, err
	}

	hasNext := strings.Contains(resp.Header.Get("Link"), `rel="next"`)

	return repos, hasNext, nil
}

// GetGitHubRepositoryLanguages extracts programming languages from a repository
func GetGitHubRepositoryLanguages(languagesURL string) ([]string, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", languagesURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	if TestOAuthToken() {
		req.Header.Set("Authorization", fmt.Sprintf("token %s", GetOAuthToken()))
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return []string{}, nil
	}

	var languageStats map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&languageStats); err != nil {
		return []string{}, nil
	}

	languages := make([]string, 0, len(languageStats))
	for lang := range languageStats {
		languages = append(languages, lang)
	}
	sort.Strings(languages)

	return languages, nil
}

// GetGitHubRepositoryLicenseURL finds the license file URL
func GetGitHubRepositoryLicenseURL(repositoryURL, branch string) string {
	urls := []string{
		fmt.Sprintf("%s/blob/%s/LICENSE", repositoryURL, branch),
		fmt.Sprintf("%s/blob/%s/LICENSE.md", repositoryURL, branch),
		fmt.Sprintf("%s/blob/%s/LICENSE.txt", repositoryURL, branch),
	}

	for _, urlStr := range urls {
		if TestURL(urlStr) {
			return urlStr
		}
	}

	return ""
}

// GetGitHubRepositoryLicense retrieves license information from GitHub
func GetGitHubRepositoryLicense(organization, repositoryURL, project, branch string) (*License, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	uri := fmt.Sprintf("%s/repos/%s/%s/license", GitHubBaseURI, strings.ToLower(organization), project)

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	if TestOAuthToken() {
		req.Header.Set("Authorization", fmt.Sprintf("token %s", GetOAuthToken()))
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var lic GitHubLicense
	if err := json.NewDecoder(resp.Body).Decode(&lic); err != nil {
		return nil, err
	}

	license := &License{}

	if lic.Message != "" || resp.StatusCode != http.StatusOK {
		license.URL = GetGitHubRepositoryLicenseURL(repositoryURL, branch)
		license.Name = ""
	} else {
		license.URL = lic.HTMLURL
		license.Name = lic.License.SPDXID
	}

	return license, nil
}

// GetGitHubRepositoryDisclaimerURL finds the disclaimer file URL
func GetGitHubRepositoryDisclaimerURL(repositoryURL, branch string) string {
	urls := []string{
		fmt.Sprintf("%s/blob/%s/DISCLAIMER", repositoryURL, branch),
		fmt.Sprintf("%s/blob/%s/DISCLAIMER.md", repositoryURL, branch),
		fmt.Sprintf("%s/blob/%s/DISCLAIMER.txt", repositoryURL, branch),
	}

	for _, urlStr := range urls {
		if TestURL(urlStr) {
			return urlStr
		}
	}

	return ""
}

// GetGitHubRepositoryReleaseURL finds the release/download URL
func GetGitHubRepositoryReleaseURL(releasesURL string) (string, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	uri := strings.Replace(releasesURL, "{/id}", "", -1)

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	if TestOAuthToken() {
		req.Header.Set("Authorization", fmt.Sprintf("token %s", GetOAuthToken()))
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", nil
	}

	var releases []GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return "", nil
	}

	for _, release := range releases {
		if !release.Prerelease {
			url := strings.Replace(release.ZipballURL, "api.", "", 1)
			return url, nil
		}
	}

	return "", nil
}

// NewCodeGovJSON generates a code.gov JSON object from GitHub data
func NewCodeGovJSON(organizations []string, agencyName, agencyEmail string, agencyOptions map[string]string, includePrivate, includeForks bool) (*CodeGovJSON, error) {
	var releases []Release

	for _, org := range organizations {
		repos, err := GetGitHubRepositories(org)
		if err != nil {
			log.Printf("Error fetching repositories for %s: %v\n", org, err)
			continue
		}

		for _, repo := range repos {
			if repo.Private != includePrivate || repo.Fork != includeForks {
				continue
			}

			release, err := buildRelease(org, repo, agencyName, agencyEmail, agencyOptions)
			if err != nil {
				log.Printf("Error building release for %s/%s: %v\n", org, repo.Name, err)
				continue
			}

			releases = append(releases, release)
		}
	}

	sort.Slice(releases, func(i, j int) bool {
		return releases[i].Name < releases[j].Name
	})

	codeGov := &CodeGovJSON{
		Version: "2.0",
		Agency:  agencyName,
		MeasurementType: MeasurementType{
			Method: "projects",
		},
		Releases: releases,
	}

	return codeGov, nil
}

func buildRelease(org string, repo GitHubRepository, agencyName, agencyEmail string, agencyOptions map[string]string) (Release, error) {
	contact := Contact{
		Email: agencyEmail,
	}

	if name, ok := agencyOptions["name"]; ok {
		contact.Name = name
	}
	if contactURL, ok := agencyOptions["url"]; ok {
		contact.URL = contactURL
	}
	if phone, ok := agencyOptions["phone"]; ok {
		contact.Phone = phone
	}

	languages, _ := GetGitHubRepositoryLanguages(repo.LanguagesURL)

	lic, err := GetGitHubRepositoryLicense(org, repo.HTMLURL, repo.Name, repo.DefaultBranch)
	if err != nil {
		lic = &License{}
	}

	disclaimerURL := GetGitHubRepositoryDisclaimerURL(repo.HTMLURL, repo.DefaultBranch)

	downloadURL, _ := GetGitHubRepositoryReleaseURL(repo.ReleasesURL)
	if downloadURL == "" {
		downloadURL = fmt.Sprintf("%s/archive/%s.zip", repo.HTMLURL, repo.DefaultBranch)
	}

	description := repo.Description
	if description == "" {
		description = "No description provided"
	}

	tags := repo.Topics
	if len(tags) == 0 {
		tags = []string{"none"}
	}

	homepageURL := repo.Homepage
	if homepageURL == "" {
		homepageURL = repo.HTMLURL
	}

	status := "Production"
	if repo.Archived {
		status = "Archival"
	}

	release := Release{
		Name:           repo.Name,
		RepositoryURL:  repo.HTMLURL,
		Description:    description,
		Permissions: Permissions{
			Licenses: []License{
				{
					URL:  lic.URL,
					Name: lic.Name,
				},
			},
			UsageType: "openSource",
		},
		LaborHours:   1,
		Tags:         tags,
		Contact:      contact,
		Status:       status,
		VCS:          "git",
		HomepageURL:  homepageURL,
		DownloadURL:  downloadURL,
		Languages:    languages,
		DisclaimerURL: disclaimerURL,
		Date: DateInfo{
			Created:             repo.CreatedAt.Format("2006-01-02"),
			LastModified:        repo.PushedAt.Format("2006-01-02"),
			MetadataLastUpdated: repo.UpdatedAt.Format("2006-01-02"),
		},
	}

	return release, nil
}

// NewCodeGovJSONFile generates and saves code.gov JSON to a file
func NewCodeGovJSONFile(organizations []string, agencyName, agencyEmail string, agencyOptions map[string]string, includePrivate, includeForks bool, outputPath string) error {
	codeGov, err := NewCodeGovJSON(organizations, agencyName, agencyEmail, agencyOptions, includePrivate, includeForks)
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(codeGov, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(outputPath, data, 0644)
}

// TestCodeGovJSONFile validates a code.gov JSON file against the schema
func TestCodeGovJSONFile(filePath string) (bool, []string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return false, nil, err
	}

	var codeGov CodeGovJSON
	if err := json.Unmarshal(data, &codeGov); err != nil {
		return false, nil, err
	}

	var errors []string

	// Basic validation
	if codeGov.Version == "" {
		errors = append(errors, "version is required")
	}
	if codeGov.Agency == "" {
		errors = append(errors, "agency is required")
	}
	if codeGov.MeasurementType.Method == "" {
		errors = append(errors, "measurementType.method is required")
	}
	if len(codeGov.Releases) == 0 {
		errors = append(errors, "releases is required and must not be empty")
	}

	for i, release := range codeGov.Releases {
		releaseErrors := validateRelease(release)
		for _, e := range releaseErrors {
			errors = append(errors, fmt.Sprintf("releases[%d]: %s", i, e))
		}
	}

	return len(errors) == 0, errors, nil
}

func validateRelease(release Release) []string {
	var errors []string

	if release.Name == "" {
		errors = append(errors, "name is required")
	}
	if release.RepositoryURL == "" {
		errors = append(errors, "repositoryURL is required")
	}
	if release.Description == "" {
		errors = append(errors, "description is required")
	}
	if len(release.Tags) == 0 {
		errors = append(errors, "tags is required")
	}
	if release.Contact.Email == "" {
		errors = append(errors, "contact.email is required")
	}
	if release.LaborHours == 0 {
		errors = append(errors, "laborHours is required and must not be 0")
	}
	if len(release.Permissions.Licenses) == 0 {
		errors = append(errors, "permissions.licenses is required")
	} else {
		for i, lic := range release.Permissions.Licenses {
			if lic.URL == "" {
				errors = append(errors, fmt.Sprintf("permissions.licenses[%d].URL is required", i))
			}
			if lic.Name == "" {
				errors = append(errors, fmt.Sprintf("permissions.licenses[%d].name is required", i))
			}
		}
	}

	return errors
}

// InvokeCodeGovJsonOverride applies overrides to a code.gov JSON file
func InvokeCodeGovJsonOverride(originalPath, newPath, overridePath string) error {
	originalData, err := os.ReadFile(originalPath)
	if err != nil {
		return err
	}

	var codeGov CodeGovJSON
	if err := json.Unmarshal(originalData, &codeGov); err != nil {
		return err
	}

	overrideData, err := os.ReadFile(overridePath)
	if err != nil {
		return err
	}

	var overrides OverrideJSON
	if err := json.Unmarshal(overrideData, &overrides); err != nil {
		return err
	}

	// Build a map of releases by name
	releaseMap := make(map[string]*Release)
	for i := range codeGov.Releases {
		releaseMap[codeGov.Releases[i].Name] = &codeGov.Releases[i]
	}

	// Apply overrides
	for _, override := range overrides.Overrides {
		release, ok := releaseMap[override.Project]
		if !ok {
			log.Printf("Release %s not found\n", override.Project)
			continue
		}

		switch override.Action {
		case "replaceproperty":
			applyReplaceProperty(release, override.Property, override.Value)
		case "addproperty":
			log.Printf("Add property not yet implemented\n")
		case "removeproperty":
			log.Printf("Remove property not yet implemented\n")
		case "removeproject":
			delete(releaseMap, override.Project)
		default:
			log.Printf("Unknown action: %s\n", override.Action)
		}
	}

	// Reconstruct releases array
	releases := make([]Release, 0, len(releaseMap))
	for _, release := range releaseMap {
		releases = append(releases, *release)
	}
	sort.Slice(releases, func(i, j int) bool {
		return releases[i].Name < releases[j].Name
	})
	codeGov.Releases = releases

	// Write output
	data, err := json.MarshalIndent(codeGov, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(newPath, data, 0644)
}

func applyReplaceProperty(release *Release, property string, value interface{}) {
	parts := strings.Split(property, ".")

	if len(parts) == 1 {
		switch property {
		case "laborHours":
			if v, ok := value.(float64); ok {
				release.LaborHours = v
			}
		}
	}
}
