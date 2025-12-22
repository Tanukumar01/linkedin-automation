package search

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"

	"github.com/Tanukumar01/linkedin-automation/internal/config"
	"github.com/Tanukumar01/linkedin-automation/internal/logger"
	"github.com/Tanukumar01/linkedin-automation/internal/stealth"
	"github.com/Tanukumar01/linkedin-automation/internal/storage"
)

// Searcher handles LinkedIn search operations
type Searcher struct {
	page     *rod.Page
	config   *config.SearchConfig
	db       *storage.DB
	timing   *stealth.TimingController
	scroller *stealth.Scroller
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
	logger.Infof("Navigating to search URL...")
	if err := s.page.Navigate(searchURL); err != nil {
		return nil, fmt.Errorf("failed to navigate to search: %w", err)
	}

	// Use a more robust wait - wait for the search results container instead of full page load
	logger.Info("Waiting for search results to appear...")
	err := s.page.Timeout(30*time.Second).WaitElementsMoreThan(".reusable-search__result-container, .entity-result", 0)
	if err != nil {
		logger.Warnf("Search results container didn't appear in 30s: %v. Continuing anyway...", err)
	}

	s.timing.Wait(s.timing.ThinkTime())

	// Take a screenshot for debugging search results
	if data, sErr := s.page.Screenshot(true, nil); sErr == nil {
		os.WriteFile("search_results_debug.png", data, 0644)
		logger.Infof("Search results screenshot saved to search_results_debug.png")
	}

	// Scroll to load results
	logger.Info("Scrolling to ensure results are loaded...")
	if err := s.scroller.ScrollDown(s.page, 800); err != nil {
		logger.Warnf("Failed to scroll: %v", err)
	}

	// Check for "No results found"
	if hasNoResults, _, _ := s.page.Has("h2.artdeco-empty-state__headline"); hasNoResults {
		logger.Warn("LinkedIn reported no results for this search.")
		return nil, nil
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
			logger.Infof("Processing found profile: %s (%s)", result.Name, result.URL)
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

	var parts []string

	// 1. Handle Job Titles (Join with OR for flexibility)
	if len(s.config.Filters.JobTitles) > 0 {
		var titles []string
		for _, t := range s.config.Filters.JobTitles {
			titles = append(titles, fmt.Sprintf("\"%s\"", t))
		}
		parts = append(parts, fmt.Sprintf("(%s)", strings.Join(titles, " OR ")))
	}

	// 2. Add basic keywords
	if len(s.config.Filters.Keywords) > 0 {
		parts = append(parts, strings.Join(s.config.Filters.Keywords, " "))
	}

	// 3. Add locations
	if len(s.config.Filters.Locations) > 0 {
		parts = append(parts, strings.Join(s.config.Filters.Locations, " "))
	}

	params := url.Values{}
	if len(parts) > 0 {
		params.Add("keywords", strings.Join(parts, " "))
	}
	params.Add("origin", "GLOBAL_SEARCH_HEADER")

	return baseURL + params.Encode()
}

// parseSearchResults parses search results from current page
func (s *Searcher) parseSearchResults() ([]ProfileResult, error) {
	// Wait for results to load and ensure page is ready
	s.timing.Wait(s.timing.ShortPause())

	// LinkedIn search results are in a list
	// Try multiple selectors as LinkedIn often AB tests layouts
	selectors := []string{
		"li.reusable-search__result-container",
		"div.search-results-container li",
		".entity-result",
	}

	var elements rod.Elements
	var err error
	for _, selector := range selectors {
		elements, err = s.page.Elements(selector)
		if err == nil && len(elements) > 0 {
			break
		}
	}

	if err != nil || len(elements) == 0 {
		return nil, fmt.Errorf("failed to find result elements: %w", err)
	}

	var results []ProfileResult

	for _, element := range elements {
		result, err := s.parseResultElement(element)
		if err != nil {
			continue
		}

		if result != nil && result.URL != "" {
			results = append(results, *result)
		}
	}

	return results, nil
}

// parseResultElement parses a single result element
func (s *Searcher) parseResultElement(element *rod.Element) (*ProfileResult, error) {
	result := &ProfileResult{}

	// Get profile URL and Name (they are usually in the same link)
	// Look for the primary title link
	linkElement, err := element.Element("a.app-aware-link")
	if err != nil {
		// Try a more generic link if specific one fails
		linkElement, err = element.Element("a[href*='/in/']")
		if err != nil {
			return nil, err
		}
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

	// Get name - often inside the link in a span
	nameElement, err := linkElement.Element("span[aria-hidden='true']")
	if err == nil {
		name, _ := nameElement.Text()
		result.Name = strings.TrimSpace(name)
	}

	// If name still empty, try looking in the whole element
	if result.Name == "" {
		if nameEl, err := element.Element(".entity-result__title-text"); err == nil {
			name, _ := nameEl.Text()
			result.Name = strings.TrimSpace(name)
		}
	}

	// Get job title
	if titleElement, err := element.Element(".entity-result__primary-subtitle"); err == nil {
		title, _ := titleElement.Text()
		result.JobTitle = strings.TrimSpace(title)
	}

	// Get location
	if locElement, err := element.Element(".entity-result__secondary-subtitle"); err == nil {
		loc, _ := locElement.Text()
		result.Location = strings.TrimSpace(loc)
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

	// Look for "Next" button - try multiple ways
	var nextButton *rod.Element
	var err error

	// Try finding by aria-label first
	nextButton, err = s.page.Element("button[aria-label*='Next']")
	if err != nil {
		// Try finding by text
		nextButton, err = s.page.ElementR("button", "(?i)Next")
	}

	if err != nil {
		return false, nil // No next button found
	}

	// Check if button is disabled
	disabled, err := nextButton.Property("disabled")
	if err == nil && disabled.Bool() {
		return false, nil
	}

	// Ensure button is in view
	nextButton.MustScrollIntoView()

	// Click next button
	if err := nextButton.Click(proto.InputMouseButtonLeft, 1); err != nil {
		return false, err
	}

	// Wait for page to load
	s.timing.Wait(s.timing.ShortPause())
	if err := s.page.WaitLoad(); err != nil {
		logger.Warnf("Failed to wait for next page load: %v", err)
	}

	return true, nil
}
