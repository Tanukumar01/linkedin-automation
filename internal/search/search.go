package search

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/go-rod/rod"
	
	"github.com/Tanukumar01/linkedin-automation/internal/config"
	"github.com/Tanukumar01/linkedin-automation/internal/logger"
	"github.com/Tanukumar01/linkedin-automation/internal/stealth"
	"github.com/Tanukumar01/linkedin-automation/internal/storage"
)

// Searcher handles LinkedIn search operations
type Searcher struct {
	page      *rod.Page
	config    *config.SearchConfig
	db        *storage.DB
	timing    *stealth.TimingController
	scroller  *stealth.Scroller
}

// ProfileResult represents a search result
type ProfileResult struct {
	URL      string
	Name     string
	JobTitle string
	Company  string
	Location string
}

// NewSearcher creates a new searcher
func NewSearcher(page *rod.Page, cfg *config.SearchConfig, db *storage.DB, timing *stealth.TimingController, scroller *stealth.Scroller) *Searcher {
	return &Searcher{
		page:     page,
		config:   cfg,
		db:       db,
		timing:   timing,
		scroller: scroller,
	}
}

// Search performs a LinkedIn search
func (s *Searcher) Search() ([]ProfileResult, error) {
	logger.Info("Starting LinkedIn search")

	// Build search URL
	searchURL := s.buildSearchURL()
	logger.Infof("Search URL: %s", searchURL)

	// Navigate to search
	if err := s.page.Navigate(searchURL); err != nil {
		return nil, fmt.Errorf("failed to navigate to search: %w", err)
	}

	if err := s.page.WaitLoad(); err != nil {
		return nil, fmt.Errorf("failed to wait for search page: %w", err)
	}

	s.timing.Wait(s.timing.ThinkTime())

	// Scroll to load results
	if err := s.scroller.ScrollDown(s.page, 500); err != nil {
		logger.Warnf("Failed to scroll: %v", err)
	}

	s.timing.Wait(s.timing.ShortPause())

	var allResults []ProfileResult
	resultsCollected := 0

	// Paginate through results
	for resultsCollected < s.config.MaxResults {
		// Parse current page
		results, err := s.parseSearchResults()
		if err != nil {
			logger.Errorf("Failed to parse search results: %v", err)
			break
		}

		if len(results) == 0 {
			logger.Info("No more results found")
			break
		}

		// Save results to database
		for _, result := range results {
			// Check if already contacted
			contacted, err := s.db.IsProfileContacted(result.URL)
			if err != nil {
				logger.Warnf("Failed to check if profile contacted: %v", err)
			}

			// Save to database
			searchResult := &storage.SearchResult{
				ProfileURL:  result.URL,
				ProfileName: result.Name,
				JobTitle:    result.JobTitle,
				Company:     result.Company,
				Location:    result.Location,
				FoundAt:     time.Now(),
				Contacted:   contacted,
			}

			if err := s.db.SaveSearchResult(searchResult); err != nil {
				logger.Warnf("Failed to save search result: %v", err)
			}
		}

		allResults = append(allResults, results...)
		resultsCollected += len(results)

		logger.Infof("Collected %d results so far", resultsCollected)

		// Check if we have enough results
		if resultsCollected >= s.config.MaxResults {
			break
		}

		// Try to go to next page
		hasNext, err := s.goToNextPage()
		if err != nil || !hasNext {
			logger.Info("No more pages available")
			break
		}

		// Random delay between pages
		delay := time.Duration(s.config.PaginationDelayMin+int(time.Now().Unix())%(s.config.PaginationDelayMax-s.config.PaginationDelayMin+1)) * time.Second
		s.timing.Wait(delay)
	}

	logger.Infof("Search completed. Total results: %d", len(allResults))

	// Log activity
	s.db.LogActivity("search", fmt.Sprintf("Found %d profiles", len(allResults)))

	return allResults, nil
}

// buildSearchURL builds the LinkedIn search URL with filters
func (s *Searcher) buildSearchURL() string {
	baseURL := "https://www.linkedin.com/search/results/people/?"

	params := url.Values{}

	// Add keywords
	if len(s.config.Filters.Keywords) > 0 {
		keywords := strings.Join(s.config.Filters.Keywords, " ")
		params.Add("keywords", keywords)
	}

	// Add job titles
	if len(s.config.Filters.JobTitles) > 0 {
		// LinkedIn uses currentJobTitle filter
		for _, title := range s.config.Filters.JobTitles {
			params.Add("title", title)
		}
	}

	// Add companies
	if len(s.config.Filters.Companies) > 0 {
		for _, company := range s.config.Filters.Companies {
			params.Add("company", company)
		}
	}

	// Add locations
	if len(s.config.Filters.Locations) > 0 {
		for _, location := range s.config.Filters.Locations {
			params.Add("geoUrn", location)
		}
	}

	return baseURL + params.Encode()
}

// parseSearchResults parses search results from current page
func (s *Searcher) parseSearchResults() ([]ProfileResult, error) {
	// Wait for results to load
	time.Sleep(2 * time.Second)

	// Find all result items
	elements, err := s.page.Elements("li.reusable-search__result-container")
	if err != nil {
		return nil, fmt.Errorf("failed to find result elements: %w", err)
	}

	var results []ProfileResult

	for _, element := range elements {
		result, err := s.parseResultElement(element)
		if err != nil {
			logger.Warnf("Failed to parse result element: %v", err)
			continue
		}

		if result != nil {
			results = append(results, *result)
		}
	}

	return results, nil
}

// parseResultElement parses a single result element
func (s *Searcher) parseResultElement(element *rod.Element) (*ProfileResult, error) {
	result := &ProfileResult{}

	// Get profile URL
	linkElement, err := element.Element("a.app-aware-link")
	if err != nil {
		return nil, err
	}

	href, err := linkElement.Property("href")
	if err != nil {
		return nil, err
	}

	result.URL = href.String()

	// Clean URL (remove query parameters)
	if idx := strings.Index(result.URL, "?"); idx != -1 {
		result.URL = result.URL[:idx]
	}

	// Get name
	nameElement, err := element.Element("span[aria-hidden='true']")
	if err == nil {
		name, _ := nameElement.Text()
		result.Name = strings.TrimSpace(name)
	}

	// Get job title
	titleElement, err := element.Element("div.entity-result__primary-subtitle")
	if err == nil {
		title, _ := titleElement.Text()
		result.JobTitle = strings.TrimSpace(title)
	}

	// Get company (usually in secondary subtitle)
	companyElement, err := element.Element("div.entity-result__secondary-subtitle")
	if err == nil {
		company, _ := companyElement.Text()
		result.Company = strings.TrimSpace(company)
	}

	// Get location
	locationElement, err := element.Element("div.entity-result__summary-metadata")
	if err == nil {
		location, _ := locationElement.Text()
		result.Location = strings.TrimSpace(location)
	}

	return result, nil
}

// goToNextPage navigates to the next page of results
func (s *Searcher) goToNextPage() (bool, error) {
	// Scroll to bottom to load pagination
	if err := s.scroller.ScrollToBottom(s.page); err != nil {
		logger.Warnf("Failed to scroll to bottom: %v", err)
	}

	s.timing.Wait(s.timing.ShortPause())

	// Look for "Next" button
	nextButton, err := s.page.Element("button[aria-label='Next']")
	if err != nil {
		return false, nil
	}

	// Check if button is disabled
	disabled, err := nextButton.Property("disabled")
	if err == nil && disabled.Bool() {
		return false, nil
	}

	// Click next button
	if err := nextButton.Click(rod.ButtonLeft, 1); err != nil {
		return false, err
	}

	// Wait for page to load
	time.Sleep(3 * time.Second)

	return true, nil
}
