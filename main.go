package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/jedib0t/go-pretty/progress"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

const (
	githubURL     = "https://github.com"
	itemSelector  = "#dependents > .Box > div[data-test-id='dg-repo-pkg-dependent']"
	repoSelector  = "a[data-hovercard-type='repository']"
	starsSelector = "div:last-child > span:nth-child(1)"
	forksSelector = "div:last-child > span:nth-child(2)"
)

type Repo struct {
	Name  string `json:"name"`
	URL   string `json:"url"`
	Stars int    `json:"stars"`
	Forks int    `json:"forks"`
}

var (
	isPackages bool
	isJSON     bool
	rows       int
	minStar    int
)

var rootCmd = &cobra.Command{
	Use:   "ghtopdep [flags] URL",
	Short: "CLI tool for sorting dependent repositories by stars",
	Args:  cobra.ExactArgs(1),
	Run:   run,
}

func init() {
	rootCmd.Flags().BoolVar(&isPackages, "packages", false, "Sort packages instead of repositories")
	rootCmd.Flags().BoolVar(&isJSON, "json", false, "Output as JSON")
	rootCmd.Flags().IntVar(&rows, "rows", 10, "Number of repositories to show in output")
	rootCmd.Flags().IntVar(&minStar, "minstar", 5, "Minimum number of stars")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) {
	url := args[0]

	repos, err := fetchDependents(url, !isPackages)
	if err != nil {
		fmt.Printf("Error fetching dependents: %v\n", err)
		os.Exit(1)
	}

	sortedRepos := sortRepos(repos, rows, minStar)

	if isJSON {
		displayJSON(sortedRepos)
	} else {
		displayTable(sortedRepos)
	}
}

func fetchDependents(url string, isRepositories bool) ([]Repo, error) {
	dependentType := "REPOSITORY"
	if !isRepositories {
		dependentType = "PACKAGE"
	}

	pageURL := fmt.Sprintf("%s/network/dependents?dependent_type=%s", url, dependentType)

	var repos []Repo
	pageCount := 0
	totalFetched := 0
	matchingStarCriteria := 0

	// Initialize progress writer
	pw := progress.NewWriter()
	pw.SetUpdateFrequency(time.Millisecond * 100)
	pw.Style().Colors = progress.StyleColorsExample

	// Start the progress writer
	go pw.Render()

	// Create a tracker for the progress bar
	tracker := &progress.Tracker{
		Message: "Fetching dependents",
		Total:   100,
		Units:   progress.UnitsDefault,
	}
	pw.AppendTracker(tracker)

	for {
		resp, err := http.Get(pageURL)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch page %s: %v", pageURL, err)
		}
		defer resp.Body.Close()

		doc, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to parse page %s: %v", pageURL, err)
		}

		pageCount++
		pageFetched := 0

		doc.Find(itemSelector).Each(func(i int, row *goquery.Selection) {
			repoElement := row.Find(repoSelector)
			name := strings.TrimSpace(repoElement.Text())
			repoURL, _ := repoElement.Attr("href")
			fullURL := githubURL + repoURL

			starsText := strings.TrimSpace(row.Find(starsSelector).Text())
			stars, _ := strconv.Atoi(strings.ReplaceAll(starsText, ",", ""))

			forksText := strings.TrimSpace(row.Find(forksSelector).Text())
			forks, _ := strconv.Atoi(strings.ReplaceAll(forksText, ",", ""))

			repos = append(repos, Repo{
				Name:  name,
				URL:   fullURL,
				Stars: stars,
				Forks: forks,
			})
			pageFetched++

			if stars >= minStar {
				matchingStarCriteria++
			}
		})

		totalFetched += pageFetched

		// Update the tracker
		tracker.SetValue(int64(totalFetched))

		// Print current status
		fmt.Printf("\rFetching dependents (Page: %d, Total: %d, Matching: %d)",
			pageCount, totalFetched, matchingStarCriteria)

		nextPage := doc.Find("#dependents > div.paginate-container > div > a:contains('Next')")
		if nextPage.Length() == 0 {
			break
		}
		pageURL, _ = nextPage.Attr("href")
	}

	// Mark the tracker as complete
	tracker.MarkAsDone()

	// Stop the progress writer
	pw.Stop()

	fmt.Printf("\nTotal dependents fetched: %d\n", totalFetched)
	fmt.Printf("Dependents matching minimum star criteria (%d): %d\n", minStar, matchingStarCriteria)

	return repos, nil
}

func sortRepos(repos []Repo, rows, minStar int) []Repo {
	sort.Slice(repos, func(i, j int) bool {
		return repos[i].Stars > repos[j].Stars
	})

	var result []Repo
	for _, repo := range repos {
		if repo.Stars >= minStar {
			result = append(result, repo)
		}
		if len(result) == rows {
			break
		}
	}

	return result
}

func displayTable(repos []Repo) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Name", "URL", "Stars", "Forks"})
	for _, repo := range repos {
		t.AppendRow([]interface{}{repo.Name, repo.URL, repo.Stars, repo.Forks})
	}
	t.SetStyle(table.StyleLight)
	t.Style().Options.SeparateRows = true
	t.Render()
}

func displayJSON(repos []Repo) {
	jsonData, err := json.MarshalIndent(repos, "", "  ")
	if err != nil {
		fmt.Printf("Error marshalling JSON: %v\n", err)
		return
	}
	fmt.Println(string(jsonData))
}
