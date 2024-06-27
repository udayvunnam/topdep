package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/spf13/cobra"
)

const (
	githubURL     = "https://github.com"
	itemSelector  = "#dependents > div.Box > div.flex-items-center"
	repoSelector  = "span > a.text-bold"
	starsSelector = "div > span:nth-child(1)"
)

type Repo struct {
	URL   string `json:"url"`
	Stars int    `json:"stars"`
}

var (
	isPackages bool
	isJSON     bool
	rows       int
	minStar    int
)

var rootCmd = &cobra.Command{
	Use:   "ghtopdep [flags] URL",
	Short: "CLI tool for sorting dependents repo by stars",
	Args:  cobra.ExactArgs(1),
	Run:   run,
}

func init() {
	rootCmd.Flags().BoolVar(&isPackages, "packages", false, "Sort packages instead of repositories")
	rootCmd.Flags().BoolVar(&isJSON, "json", false, "Output as JSON")
	rootCmd.Flags().IntVar(&rows, "rows", 10, "Number of repositories to show")
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

	repos := []Repo{}
	for {
		doc, err := goquery.NewDocument(pageURL)
		if err != nil {
			return nil, err
		}

		doc.Find(itemSelector).Each(func(i int, s *goquery.Selection) {
			starsText := strings.TrimSpace(s.Find(starsSelector).Text())
			stars, _ := strconv.Atoi(strings.ReplaceAll(starsText, ",", ""))

			repoURL, _ := s.Find(repoSelector).Attr("href")
			fullURL := githubURL + repoURL

			repos = append(repos, Repo{URL: fullURL, Stars: stars})
		})

		nextPage := doc.Find("#dependents > div.paginate-container > div > a:contains('Next')")
		if nextPage.Length() == 0 {
			break
		}
		pageURL, _ = nextPage.Attr("href")
	}

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
	fmt.Println("| URL | Stars |")
	fmt.Println("|-----|-------|")
	for _, repo := range repos {
		fmt.Printf("| %s | %d |\n", repo.URL, repo.Stars)
	}
}

func displayJSON(repos []Repo) {
	jsonData, _ := json.MarshalIndent(repos, "", "  ")
	fmt.Println(string(jsonData))
}
