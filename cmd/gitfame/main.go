package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	flag "github.com/spf13/pflag"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
)

type L struct {
	Name       string
	Type       string
	Extensions []string
}

func GitLog(rep, commit, fileName string) (string, string) {
	// Construct the git log command with the desired format
	cmd := exec.Command("git", "log", "--pretty=format:%H %an", commit, "--", fileName)
	cmd.Dir = rep // Set the working directory to the repository path

	// Execute the command and capture the output
	b, err := cmd.Output()
	if err != nil {
		log.Fatalf("Failed to execute git log command: %v", err)
	}

	// Convert the output bytes to a string
	output := string(b)
	// Split the output into lines
	lines := strings.Split(output, "\n")
	if len(lines) == 0 {
		log.Fatal("No output from git log command")
	}

	// Split the first line into commit hash and author name
	parts := strings.SplitN(lines[0], " ", 2)
	if len(parts) < 2 {
		log.Fatal("Unexpected format of git log output")
	}

	// Return the commit hash and author name
	return parts[0], parts[1]
}

func GitBlame(rep, commit, name string) []string {
	cmd := exec.Command("git", "blame", "--porcelain", commit, name)
	cmd.Dir = rep
	b, _ := cmd.Output()

	var s strings.Builder
	s.Write(b)
	return strings.FieldsFunc(s.String(), func(r rune) bool {
		return r == '\n'
	})
}

func GitLsTree(rep, commit string, extensions, languages []string) []string {
	cmd := exec.Command("git", "ls-tree", "-r", "--name-only", commit)
	cmd.Dir = rep
	b, err := cmd.Output()
	if err != nil {
		panic(err)
	}

	var s strings.Builder
	s.Write(b)
	out := strings.FieldsFunc(s.String(), func(r rune) bool {
		return r == '\n'
	})

	ans := out

	if len(languages) != 0 {
		getLanguagesExtensions(languages, &extensions)
	}

	if len(extensions) != 0 {
		e := make(map[string]struct{}, len(extensions))
		for _, i := range extensions {
			e[i] = struct{}{}
		}

		ans = make([]string, 0)
		for _, i := range out {
			if _, ok := e[filepath.Ext(i)]; ok {
				ans = append(ans, i)
			}
		}
	}

	return ans
}

func getLanguagesExtensions(languages []string, extensions *[]string) {
	b, _ := os.ReadFile("../../configs/language_extensions.json")

	var l []L
	err := json.Unmarshal(b, &l)
	if err != nil {
		panic(err)
	}

	set := make(map[string]int, len(l))

	for i, j := range l {
		set[strings.ToLower(j.Name)] = i
	}

	for _, i := range languages {
		if itr, ok := set[strings.ToLower(i)]; ok {
			*extensions = append(*extensions, l[itr].Extensions...)
		}
	}
}

type Config struct {
	Repository   string
	Commit       string
	OrderBy      string
	UseCommitter bool
	Format       string
	Extensions   []string
	Languages    []string
	Exclude      []string
	RestrictTo   []string
}

func NewConfig() Config {
	set := flag.NewFlagSet("flag_set", flag.ExitOnError)

	var cfg Config

	set.StringVar(&cfg.Repository, "repository", ".", "Specify the path to the Git repository. Uses the current directory if not specified.")
	set.StringVar(&cfg.Commit, "revision", "HEAD", "Set the commit hash or branch to analyze. Defaults to the HEAD of the current branch.")
	set.StringVar(&cfg.OrderBy, "order-by", "lines", "Determine the sorting criterion of the output. Options are 'lines', 'commits', or 'files'. Defaults to 'lines'.")
	set.BoolVar(&cfg.UseCommitter, "use-committer", false, "Use the committer instead of the author for generating statistics. Defaults to false, using the author.")
	set.StringVar(&cfg.Format, "format", "tabular", "Specify the format of the output. Options are 'tabular', 'csv', 'json', or 'json-lines'. Defaults to 'tabular'.")
	set.StringSliceVar(&cfg.Extensions, "extensions", []string{}, "Provide a list of file extensions to include in the analysis. Separate multiple extensions with commas. If empty, all extensions are included.")
	set.StringSliceVar(&cfg.Languages, "languages", []string{}, "Specify a list of programming languages to include in the analysis. This filters files based on common extensions for the specified languages. Separate multiple languages with commas.")
	set.StringSliceVar(&cfg.Exclude, "exclude", []string{}, "Define a set of glob patterns to exclude files from the analysis. Separate multiple patterns with commas.")
	set.StringSliceVar(&cfg.RestrictTo, "restrict-to", []string{}, "Define a set of glob patterns to restrict the analysis to specific files. Separate multiple patterns with commas.")

	err := set.Parse(os.Args[1:])
	if err != nil {
		log.Fatalf("Error parsing flags: %v", err)
	}

	return cfg
}

type AuthorJSON struct {
	Name    string `json:"name"`
	Lines   int    `json:"lines"`
	Commits int    `json:"commits"`
	Files   int    `json:"files"`
}

func FormatData(a AuthorSlice, format string) {
	switch format {
	case "tabular":
		formatTab(a)
	case "csv":
		formatCSV(a)
	case "json":
		formatJSON(a)
	case "json-lines":
		formatJSONlines(a)
	default:
		panic("format flag value is incorrect")
	}
}

func formatTab(a AuthorSlice) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.TabIndent)
	fmt.Fprintln(w, "Name\tLines\tCommits\tFiles")

	for _, author := range a.Slice {
		fmt.Fprintf(w, "%s\t%d\t%d\t%d\n", author.Name, author.Statistics.Lines, author.Statistics.Commits, author.Statistics.Files)
	}
	w.Flush()
}

func formatCSV(a AuthorSlice) {
	w := csv.NewWriter(os.Stdout)
	err := w.Write([]string{"Name", "Lines", "Commits", "Files"})
	if err != nil {
		panic(err)
	}

	for _, i := range a.Slice {
		err = w.Write([]string{i.Name,
			strconv.Itoa(i.Statistics.Lines),
			strconv.Itoa(i.Statistics.Commits),
			strconv.Itoa(i.Statistics.Files)},
		)
		if err != nil {
			panic(err)
		}
	}

	w.Flush()
}

func formatJSON(a AuthorSlice) {
	flatAuthors := make([]AuthorJSON, len(a.Slice))
	for i, author := range a.Slice {
		flatAuthors[i] = AuthorJSON{
			Name:    author.Name,
			Commits: author.Statistics.Commits,
			Files:   author.Statistics.Files,
			Lines:   author.Statistics.Lines,
		}
	}
	b, err := json.Marshal(flatAuthors)
	if err != nil {
		panic(err)
	}
	os.Stdout.Write(b)
}

func formatJSONlines(a AuthorSlice) {
	for _, i := range a.Slice {
		b, _ := json.Marshal(AuthorJSON{
			Name:    i.Name,
			Lines:   i.Statistics.Lines,
			Commits: i.Statistics.Commits,
			Files:   i.Statistics.Files,
		})
		os.Stdout.Write(b)
		fmt.Println()
	}

}

type Statistics struct {
	Lines   int
	Commits int
	Files   int
}

type Author struct {
	Statistics Statistics
	Name       string
}

type AuthorSlice struct {
	Slice   []Author
	orderBy string
}

func (a AuthorSlice) Len() int {
	return len(a.Slice)
}

func (a AuthorSlice) Swap(i, j int) {
	a.Slice[i], a.Slice[j] = a.Slice[j], a.Slice[i]
}

func (a AuthorSlice) Less(i, j int) bool {
	compareSlices := func(s1, s2 []int) bool {
		for k := range s1 {
			if s1[k] != s2[k] {
				return s1[k] > s2[k]
			}
		}
		return false
	}

	var key1, key2 []int

	switch a.orderBy {
	case "lines":
		key1 = []int{a.Slice[i].Statistics.Lines, a.Slice[i].Statistics.Commits, a.Slice[i].Statistics.Files}
		key2 = []int{a.Slice[j].Statistics.Lines, a.Slice[j].Statistics.Commits, a.Slice[j].Statistics.Files}
	case "commits":
		key1 = []int{a.Slice[i].Statistics.Commits, a.Slice[i].Statistics.Lines, a.Slice[i].Statistics.Files}
		key2 = []int{a.Slice[j].Statistics.Commits, a.Slice[j].Statistics.Lines, a.Slice[j].Statistics.Files}
	case "files":
		key1 = []int{a.Slice[i].Statistics.Files, a.Slice[i].Statistics.Lines, a.Slice[i].Statistics.Commits}
		key2 = []int{a.Slice[j].Statistics.Files, a.Slice[j].Statistics.Lines, a.Slice[j].Statistics.Commits}
	default:
		panic("Unsupported order-by flag value")
	}
	if compare := compareSlices(key1, key2); compare {
		return true
	} else if !compare && key1[0] == key2[0] && key1[1] == key2[1] && key1[2] == key2[2] {
		return strings.ToLower(a.Slice[i].Name) < strings.ToLower(a.Slice[j].Name)
	}
	return false
}

func restr(files, restrictTo []string) []string {
	var matchedFiles []string

	for _, file := range files {
		matchesPattern := false

		for _, pattern := range restrictTo {
			match, err := filepath.Match(pattern, file)
			if err != nil {
				fmt.Println("Error matching pattern:", err)
				continue
			}
			if match {
				matchesPattern = true
				break
			}
		}

		if matchesPattern {
			matchedFiles = append(matchedFiles, file)
		}
	}

	return matchedFiles
}

func Exclude(files, exclude []string) []string {
	var filteredFiles []string

	for _, file := range files {
		excluded := false

		for _, pattern := range exclude {
			match, err := filepath.Match(pattern, file)
			if err != nil {
				fmt.Println("Error matching pattern:", err)
				continue
			}
			if match {
				excluded = true
				break
			}
		}

		if !excluded {
			filteredFiles = append(filteredFiles, file)
		}
	}

	return filteredFiles
}

func Blame(out []string, useCommitter bool, rep, commit, fileName string) (map[string][]string, map[string]int) {
	authors := make(map[string][]string)
	commits := make(map[string]int)

	if len(out) == 0 {
		hash, a := GitLog(rep, commit, fileName)
		commits[hash] = 0
		authors[a] = append(authors[a], hash)
	}

	isNextHash := true
	itr := 0
	var isWaitForAuthor bool
	var lastHash string
	for _, i := range out {
		if isNextHash {
			isNextHash = false
			if itr == 0 {
				s := strings.Split(i, " ")
				itr, _ = strconv.Atoi(s[len(s)-1])
				commits[s[0]] += itr
				isWaitForAuthor = true
				lastHash = s[0]
			}
			itr--
		} else if i[0] != '\t' && isWaitForAuthor {
			s := strings.Split(i, " ")
			if !useCommitter && s[0] == "author" {
				name := i[len("author "):]
				authors[name] = append(authors[name], lastHash)
				isWaitForAuthor = false
			} else if useCommitter && s[0] == "committer" {
				name := i[len("committer "):]
				authors[name] = append(authors[name], lastHash)
				isWaitForAuthor = false
			}
		} else if i[0] == '\t' {
			isNextHash = true
		}
	}

	return authors, commits
}

func AuthorData(authors map[string]*Statistics, a map[string]map[string]struct{}, c, files map[string]int) {
	for author, commits := range a {
		totalLines := 0
		for commit := range commits {
			totalLines += c[commit]
		}
		authors[author] = &Statistics{
			Lines:   totalLines,
			Commits: len(commits),
			Files:   files[author],
		}
	}
}

func Sort(authors map[string]*Statistics, orderBy string) AuthorSlice {
	var authorSlice AuthorSlice
	authorSlice.orderBy = orderBy

	for name, stats := range authors {
		authorSlice.Slice = append(authorSlice.Slice, Author{
			Statistics: *stats,
			Name:       name,
		})
	}

	sort.Sort(authorSlice)
	return authorSlice
}

func main() {
	cfg := NewConfig()

	// Получение списка файлов с учетом фильтров расширений и языков программирования
	files := GitLsTree(cfg.Repository, cfg.Commit, cfg.Extensions, cfg.Languages)

	// Применение фильтров исключения
	if len(cfg.Exclude) != 0 {
		files = Exclude(files, cfg.Exclude)
	}

	// Применение фильтров ограничения
	if len(cfg.RestrictTo) != 0 {
		files = restr(files, cfg.RestrictTo)
	}

	// Инициализация структур для сбора данных
	authors := make(map[string]*Statistics)
	a1 := make(map[string]map[string]struct{}, 100)
	c1 := make(map[string]int)
	f1 := make(map[string]int)

	// Обработка файлов для сбора статистики
	for _, fileName := range files {
		blameOutput := GitBlame(cfg.Repository, cfg.Commit, fileName)
		a, c := Blame(blameOutput, cfg.UseCommitter, cfg.Repository, cfg.Commit, fileName)
		UpdateData(a1, c1, f1, a, c)
	}

	// Агрегация собранных данных
	AuthorData(authors, a1, c1, f1)

	// Сортировка авторов в соответствии с указанным критерием
	sortedAuthors := Sort(authors, cfg.OrderBy)

	// Форматирование и вывод результатов
	FormatData(sortedAuthors, cfg.Format)
}

func UpdateData(a1 map[string]map[string]struct{}, c1, f1 map[string]int, a map[string][]string, c map[string]int) {
	for key, value := range c {
		c1[key] += value
	}

	for key, values := range a {
		f1[key]++
		if _, exists := a1[key]; !exists {
			a1[key] = make(map[string]struct{})
		}
		for _, val := range values {
			a1[key][val] = struct{}{}
		}
	}
}
