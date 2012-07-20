package main

import (
	"os"
	"fmt"
	"mime"
	"sort"
	"sync"
	"time"
	"regexp"
	"strings"
	"./minheap"
	"net/url"
	"net/http"
	"crypto/md5"
	"./levenshtein"
	"encoding/json"
	"path/filepath"
	"text/template"
	"github.com/bemasher/errhandler"
)

var (
	file_re *regexp.Regexp
	config Config
)

const (
	TEMPLATE = "queue.txt"
	
	SEARCH_URL = "http://api.trakt.tv/search/shows.json/%s/%s"
	SEASONS_URL = "http://api.trakt.tv/show/seasons.json/%s/%s"
	EPISODES_URL = "http://api.trakt.tv/show/season.json/%s/%s/%d"
)

type Config struct {
	TraktAPI string
	MediaPath string
	FeedPath string
	Host string
	MediaURL string
}

// Type for storing information about each item
type Item struct {
	Filename string
	Length int64
	Type string
	Query string
	Show *Show
	SeasonNum int
	EpisodeNum int
}

// Helper to simplify retrieving the episode of an item
func (i Item) Episode() Episode {
	return i.Show.Episodes[i.SeasonNum][i.EpisodeNum]
}

// Return the S##E## code for this item
func (i Item) EpCode() string {
	return fmt.Sprintf("S%02dE%02d", i.SeasonNum, i.EpisodeNum)
}

// Return the episode's title
func (i Item) Title() string {
	return i.Episode().Title
}

// Return the episode's overview for the description field of the item
func (i Item) Description() string {
	return i.Episode().Overview
}

// Parse the airdate of the episode for use as the item's pubdate
func (i Item) PubDate() string {
	return time.Unix(i.Episode().Aired, 0).UTC().Format(time.RFC1123Z)
}

// Compute a unique hash for this item
func (i Item) GUID() string {
	md5 := md5.New()
	fmt.Fprintf(md5, "%s - %s - %s - %d", i.Query, i.EpCode(), i.Title(), i.Length)
	return fmt.Sprintf("%X", md5.Sum(nil))
}

// Implement the sort interface for a list of items so we can
// sort them by date later
type ItemList []Item

func (il ItemList) Len() int {
	return len(il)
}

func (il ItemList) Less(i, j int) bool {
	return il[i].Episode().Aired > il[j].Episode().Aired
}

func (il ItemList) Swap(i, j int) {
	il[i], il[j] = il[j], il[i]
}

// Type for storing information for each show
// and episodes belonging to it
type Show struct {
	Title string
	ID string `json:"tvdb_id"`
	Episodes map[int]map[int]Episode `json:"-"`
}

// Type for storing information of each episode
type Episode struct {
	Episode int
	Aired int64 `json:"first_aired"`
	Title string
	Overview string
}

func (s *Show) Populate(seasons map[int]bool) {
	wg := new(sync.WaitGroup)
	s.Episodes = make(map[int]map[int]Episode, 0)
	
	// For each season we have files for
	for season, _ := range seasons {
		wg.Add(1)
		// Get all the episodes for the current season
		go func(season int, wg *sync.WaitGroup) {
			r, err := http.Get(fmt.Sprintf(EPISODES_URL, config.TraktAPI, s.ID, season))
			errhandler.Handle("Error opening episodes file: ", err)
			defer r.Body.Close()
			
			var episodes []Episode
			decoder := json.NewDecoder(r.Body)
			decoder.Decode(&episodes)
			
			// Store episodes into episode map of the current show
			// for easy retrieval later
			for _, episode := range episodes {
				if _, exists := s.Episodes[season]; !exists {
					s.Episodes[season] = make(map[int]Episode, 0)
				}
				s.Episodes[season][episode.Episode] = episode
			}
			wg.Done()
		}(season, wg)
	}
	
	// Block until all episodes have been gotten
	wg.Wait()
}

func Search(query string) *Show {
	query_url := fmt.Sprintf(SEARCH_URL, config.TraktAPI, url.QueryEscape(query))
	
	// Get the search results
	r, err := http.Get(query_url)
	errhandler.Handle("Error retrieving query: ", err)
	defer r.Body.Close()
	
	searchDecoder := json.NewDecoder(r.Body)
	
	// Decode the search results
	var results []Show
	var heap minheap.Heap
	searchDecoder.Decode(&results)
	
	// Push all results onto the minheap with their levenshtein
	// scores as the priorty
	for _, s := range results {
		heap.Push(levenshtein.Levenshtein(query, s.Title), s)
	}
	
	// Get the smallest item from the heap, it's the closest match
	result := heap.Pop().(Show)
	return &result
}

// Parses episode code and show name from the filename
func ParseFilename(filename string) (show string, season int, episode int) {
	matches := file_re.FindStringSubmatch(filename)
	show, epcode := strings.ToLower(matches[1]), matches[2]
	
	show = strings.Replace(show, ".", " ", -1)
	fmt.Sscanf(epcode, "S%dE%d", &season, &episode)
	
	return 
}

func init() {
	// Regex for matching filenames of tv show episodes
	file_re = regexp.MustCompilePOSIX("(.*?)_(.*?).(mp4|avi|mkv)")
}

func main() {
	configFile, err := os.Open("config.json")
	errhandler.Handle("Error opening config file: ", err)
	defer configFile.Close()
	
	configDecoder := json.NewDecoder(configFile)
	err = configDecoder.Decode(&config)
	errhandler.Handle("Error decoding config file: ", err)
	
	globMask := filepath.Join(config.MediaPath, "*.*")
	files, err := filepath.Glob(globMask)
	errhandler.Handle("Error globbing media directory for media: ", err)
	
	showMap := make(map[string]*Show, 0)
	seasonMap := make(map[string]map[int]bool, 0)
	
	var items ItemList
	
	// For each file which matches the glob mask
	for _, file := range files {
		// Determine if it's a tv show or not
		filename := filepath.Base(file)
		if file_re.MatchString(filename) {
			// Parse out the show name, season and episode numbers
			show, season, episode := ParseFilename(filename)
			
			// Stat the file to determine length
			fileInfo, err := os.Stat(file)
			errhandler.Handle("Error statting file: ", err)
			
			// Create a new RSS feed item for the file
			items = append(items, Item{
				filename,
				fileInfo.Size(),
				mime.TypeByExtension(filepath.Ext(filename)),
				show,
				// This is a pointer to the show this episode
				// belongs to, we'll populate it later
				nil,
				season,
				episode,
			})
			
			// Make a map for getting necessary shows later
			showMap[show] = nil
			
			// Make a map of episodes in seasons we have for each show
			if _, exists := seasonMap[show]; !exists {
				seasonMap[show] = make(map[int]bool, 0)
			}
			seasonMap[show][season] = true
		}
	}
	
	wg := new(sync.WaitGroup)
	
	// Get all show and episode information necessary
	for show, _ := range showMap {
		wg.Add(1)
		go func(show string, wg *sync.WaitGroup) {
			showMap[show] = Search(show)
			showMap[show].Populate(seasonMap[show])
			wg.Done()
		}(show, wg)
	}
	
	// Block until all the information has been retrieved
	wg.Wait()
	
	// Now that we have all the information we need
	// set the show for each file
	for i, _ := range items {
		items[i].Show = showMap[items[i].Query]
	}
	
	// Sort by date, newest to oldest
	sort.Sort(items)
	
	funcMap := template.FuncMap{
		"Time": func() string {return time.Now().UTC().Format(time.RFC1123Z)},
		"Host": func() string {return config.Host},
		"MediaURL": func() string {return config.MediaURL},
	}
	
	// Parse the queue template
	t, err := template.New(TEMPLATE).Funcs(funcMap).ParseFiles(TEMPLATE)
	errhandler.Handle("Error parsing template: ", err)
	
	// Create the queue file
	rssFile, err := os.Create(config.FeedPath)
	errhandler.Handle("Error creating rss file: ", err)
	defer rssFile.Close()
	
	// Execute the template, write output to the queue file
	err = t.Execute(rssFile, items)
	errhandler.Handle("Error executing template: ", err)
}