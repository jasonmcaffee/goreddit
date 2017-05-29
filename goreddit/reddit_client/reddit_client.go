package reddit_client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

var redditClient *http.Client

func init() {
	fmt.Println("creating reddit client")
	redditClient = &http.Client{
		Timeout: 60 * time.Second,
		Transport: &http.Transport{
			MaxIdleConnsPerHost: 20,
		},
	}
}

//reddit json response types
type (
	RedditSubredditPostsResultContainer struct {
		Data RedditPostResult
	}
	RedditPostResult struct {
		Children []RedditPostContainer
		After    string
		Before   string
	}
	RedditPostContainer struct {
		Data RedditPost
	}
	RedditPost struct {
		Title string
		Url   string
		ID    string
	}

	//get back an array of these, but only 1
	RedditCommentResponse struct {
		Data RedditCommentsContainer `json:"data"`
		Kind string                  `json:"kind"`
	}
	RedditCommentsContainer struct {
		Children []RedditCommentContainer `json:"children"`
	}
	RedditCommentContainer struct {
		Data RedditComment `json:"data"`
		Kind string        `json:"kind"`
	}
	RedditComment struct {
		Comment string `json:"body"`
		ID      string `json:"id"`
	}
)

func FetchCommentsForPost(postID string, limit int) ([]RedditCommentContainer, error) {
	//comments api doesn't return the amount you ask for. ask for more then return subset
	askLimit := limit*2 + 20
	//fmt.Println("\n ......fetching comments...... ")
	url := fmt.Sprintf("https://www.reddit.com/r/all/comments/%s/top.json?sr_detail=false&limit=%s", postID, strconv.Itoa(askLimit))
	//redditClient := &http.Client{Timeout: 60 * time.Second}
	request, _ := http.NewRequest("GET", url, nil)
	request.Header.Add("Content-Type", "application/json")
	r, err := redditClient.Do(request)
	results := []RedditCommentResponse{}
	if err != nil {
		return []RedditCommentContainer{}, err
	}
	defer r.Body.Close()

	if r.StatusCode == 429 {
		//fmt.Println("done fetching. status code 429 (rate limiting by reddit) retrying shortly...: ", r.StatusCode)
		time.Sleep(500 * time.Millisecond)
		return FetchCommentsForPost(postID, limit)
	}

	responseBody, _ := ioutil.ReadAll(r.Body)
	err = json.Unmarshal(responseBody, &results)
	//fmt.Println(err)

	redditCommentContainers := []RedditCommentContainer{}
	if len(results) > 0 {
		//don't include the kind="more".. it's weird reddit mixes types together in a single array
		for i, child := range results[1].Data.Children {
			if i >= limit {
				break
			}
			if child.Kind == "t1" { //comments are type t1
				redditCommentContainers = append(redditCommentContainers, child)
			}
		}
	}

	return redditCommentContainers, err
}

func FetchSubredditPosts(after string, before string, limit int) (resp RedditSubredditPostsResultContainer, nextAfter string, nextBefore string, err error) {
	//fmt.Println("\n ......fetching posts...... ")
	url := fmt.Sprintf("https://www.reddit.com/r/all/top.json?limit=%s&after=%s&before=%s", strconv.Itoa(limit), after, before)
	request, _ := http.NewRequest("GET", url, nil)
	request.Header.Add("Content-Type", "application/json")

	result := &RedditSubredditPostsResultContainer{}
	r, err := redditClient.Do(request)
	if err != nil {
		return *result, "", "", err
	}
	defer r.Body.Close()

	//reddit does rate limiting, so recursively call again in 1 second
	if r.StatusCode == 429 {
		//fmt.Println("done fetching. status code 429 (rate limiting by reddit) retrying shortly...: ", r.StatusCode)
		time.Sleep(500 * time.Millisecond)
		return FetchSubredditPosts(after, before, limit)
	}
	responseBody, _ := ioutil.ReadAll(r.Body)
	err = json.Unmarshal(responseBody, &result)
	return *result, result.Data.After, result.Data.Before, err
}
