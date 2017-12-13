package main

import (
	"encoding/json"
	"sort"
	"log"
	"strings"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/iced-mocha/shared/models"
	"github.com/mmcdole/gofeed"
	"github.com/patrickmn/go-cache"
)

const (
	defaultPostCount = 20
	port             = ":9000"
	baseURL          = "http://rss-client" + port
)

type ByTime []models.Post

func (p ByTime) Len() int {
	return len(p)
}

func (p ByTime) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p ByTime) Less(i, j int) bool {
	return p[j].Date.Before(p[i].Date)
}

func getFeedPosts(feed *gofeed.Feed) []models.Post {
	var posts []models.Post
	items := feed.Items
	for _, item := range items {
		posts = append(posts, models.Post{
			Date:    *item.PublishedParsed,
			Author:  item.Author.Name,
			Title:   item.Title,
			Content: item.Description,
			PostLink: item.Link,
		})
	}
	return posts
}

func GetPosts(w http.ResponseWriter, r *http.Request, c *cache.Cache, id func() string, rss *gofeed.Parser) {
	var err error

//	postCountToReturn := defaultPostCount
	if countStr := r.FormValue("count"); countStr != "" {
		count, _ := strconv.Atoi(countStr)
		if count != 0 {
//			postCountToReturn = count
		}
	}

	// TODO: Do something with this
//	pagingToken := r.FormValue("continue")

	feedUrlsRaw := r.FormValue("feeds")
	if feedUrlsRaw == "" {
		http.Error(w, "Param 'feeds' required", http.StatusNotFound)
	}

	feedUrls := strings.Split(feedUrlsRaw, ",")

	var postsToReturn []models.Post
	for _, url := range feedUrls {
		feed, err := rss.ParseURL(url)
		if err != nil {
			continue
		}
		postsToReturn = append(postsToReturn, getFeedPosts(feed)...)
	}
	sort.Sort(ByTime(postsToReturn))

	// TODO: Pagination
	var nextURL string
	cRes := models.ClientResp{postsToReturn, nextURL}
	res, err := json.Marshal(cRes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(res)
}

func main() {
	var idCounter int32
	id := func() string {
		idCounter++
		return strconv.FormatInt(int64(idCounter), 32)
	}
	c := cache.New(30*time.Minute, 45*time.Minute)
	parser := gofeed.NewParser()

	f := func(w http.ResponseWriter, r *http.Request) {
		GetPosts(w, r, c, id, parser)
	}

	r := mux.NewRouter()
	r.HandleFunc("/v1/posts", f).Methods(http.MethodGet)
	log.Printf("starting server on port " + port)
	log.Fatal(http.ListenAndServe(port, r))
}
