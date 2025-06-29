package robots

import (
	"bufio"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

type RobotsTxt struct {
	UserAgentRules map[string]*UserAgentRules
	CrawlDelay     time.Duration
	SitemapURLs    []string
}

type UserAgentRules struct {
	Allow      []string
	Disallow   []string
	CrawlDelay time.Duration
}

type RobotsChecker struct {
	cache     map[string]*RobotsTxt
	mu        sync.RWMutex
	userAgent string
}

func NewRobotsChecker(userAgent string) *RobotsChecker {
	if userAgent == "" {
		userAgent = "*"
	}
	return &RobotsChecker{
		cache:     make(map[string]*RobotsTxt),
		userAgent: userAgent,
	}
}

func (rc *RobotsChecker) IsAllowed(targetURL string) (bool, time.Duration) {
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return true, 0 // If we can't parse the URL, allow it
	}

	domain := parsedURL.Scheme + "://" + parsedURL.Host
	robotsTxt := rc.getRobotsTxt(domain)

	if robotsTxt == nil {
		return true, 0 // If no robots.txt or error fetching, allow crawling
	}

	// Check user agent specific rules first, then fall back to *
	userAgents := []string{rc.userAgent, "*"}
	for _, ua := range userAgents {
		if rules, exists := robotsTxt.UserAgentRules[ua]; exists {
			allowed := rc.checkRules(parsedURL.Path, rules)
			crawlDelay := rules.CrawlDelay
			if crawlDelay == 0 {
				crawlDelay = robotsTxt.CrawlDelay
			}
			return allowed, crawlDelay
		}
	}

	return true, robotsTxt.CrawlDelay
}

func (rc *RobotsChecker) getRobotsTxt(domain string) *RobotsTxt {
	rc.mu.RLock()
	if robotsTxt, exists := rc.cache[domain]; exists {
		rc.mu.RUnlock()
		return robotsTxt
	}
	rc.mu.RUnlock()

	// Fetch and parse robots.txt
	robotsTxt := rc.fetchAndParseRobotsTxt(domain)

	rc.mu.Lock()
	rc.cache[domain] = robotsTxt
	rc.mu.Unlock()

	return robotsTxt
}

func (rc *RobotsChecker) fetchAndParseRobotsTxt(domain string) *RobotsTxt {
	robotsURL := domain + "/robots.txt"

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(robotsURL)
	if err != nil {
		fmt.Printf("Error fetching robots.txt for %s: %v\n", domain, err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Printf("robots.txt not found for %s (status: %d)\n", domain, resp.StatusCode)
		return nil
	}

	robotsTxt := &RobotsTxt{
		UserAgentRules: make(map[string]*UserAgentRules),
	}

	scanner := bufio.NewScanner(resp.Body)
	var currentUserAgent string
	var currentRules *UserAgentRules

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Split on first colon
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		directive := strings.ToLower(strings.TrimSpace(parts[0]))
		value := strings.TrimSpace(parts[1])

		switch directive {
		case "user-agent":
			currentUserAgent = strings.ToLower(value)
			if _, exists := robotsTxt.UserAgentRules[currentUserAgent]; !exists {
				robotsTxt.UserAgentRules[currentUserAgent] = &UserAgentRules{
					Allow:    make([]string, 0),
					Disallow: make([]string, 0),
				}
			}
			currentRules = robotsTxt.UserAgentRules[currentUserAgent]

		case "allow":
			if currentRules != nil && value != "" {
				currentRules.Allow = append(currentRules.Allow, value)
			}

		case "disallow":
			if currentRules != nil {
				currentRules.Disallow = append(currentRules.Disallow, value)
			}

		case "crawl-delay":
			if delay, err := strconv.Atoi(value); err == nil {
				crawlDelay := time.Duration(delay) * time.Second
				if currentRules != nil {
					currentRules.CrawlDelay = crawlDelay
				} else {
					robotsTxt.CrawlDelay = crawlDelay
				}
			}

		case "sitemap":
			robotsTxt.SitemapURLs = append(robotsTxt.SitemapURLs, value)
		}
	}

	return robotsTxt
}

func (rc *RobotsChecker) checkRules(path string, rules *UserAgentRules) bool {
	// Check Allow rules first (more specific)
	for _, allowPattern := range rules.Allow {
		if rc.matchesPattern(path, allowPattern) {
			return true
		}
	}

	// Check Disallow rules
	for _, disallowPattern := range rules.Disallow {
		if rc.matchesPattern(path, disallowPattern) {
			return false
		}
	}

	// If no rules match, allow by default
	return true
}

func (rc *RobotsChecker) matchesPattern(path, pattern string) bool {
	if pattern == "" {
		return false
	}

	// Handle wildcard at the end
	if strings.HasSuffix(pattern, "*") {
		prefix := pattern[:len(pattern)-1]
		return strings.HasPrefix(path, prefix)
	}

	// Exact match or prefix match
	return path == pattern || strings.HasPrefix(path, pattern)
}

// GetCrawlDelay returns the crawl delay for a specific URL
func (rc *RobotsChecker) GetCrawlDelay(targetURL string) time.Duration {
	_, delay := rc.IsAllowed(targetURL)
	return delay
}
