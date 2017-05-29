package reddit_client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

const (
	rateLimitWaitMs time.Duration = 200
)
var (
	redditClient *http.Client
)

//when this module is loaded, create an http client which allows for multiple requests to be sent at the same time.
func init() {
	//fmt.Println("creating reddit client")
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
		SelfText string //e.g. askreddit question, joke punchline, etc
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

//FetchCommentsForPost fetches comments for the given post id.
//Reddit's api uses rate limiting, so any given request can have a 429 response
//if 429 is encountered, this function will sleep the thread for N milliseconds, then recursively try again
func FetchCommentsForPost(postID string, limit int) ([]RedditCommentContainer, error) {
	//comments api doesn't return the amount you ask for. ask for more then return subset
	askLimit := limit * 2 + 20
	//fmt.Println("\n ......fetching comments...... ")
	url := fmt.Sprintf("https://www.reddit.com/r/all/comments/%s/top.json?sr_detail=false&limit=%s", postID, strconv.Itoa(askLimit))
	request, _ := http.NewRequest("GET", url, nil)
	request.Header.Add("Content-Type", "application/json")
	r, err := redditClient.Do(request)
	results := []RedditCommentResponse{}
	if err != nil {
		return []RedditCommentContainer{}, err
	}
	defer r.Body.Close() //close the connection once func has exited
	//handle reddit's rate limiting
	if r.StatusCode == 429 {
		//fmt.Println("done fetching. status code 429 (rate limiting by reddit) retrying shortly...: ", r.StatusCode)
		time.Sleep(rateLimitWaitMs * time.Millisecond)
		return FetchCommentsForPost(postID, limit)
	}
	//deserialize
	responseBody, _ := ioutil.ReadAll(r.Body)
	err = json.Unmarshal(responseBody, &results)
	if err != nil {
		return []RedditCommentContainer{}, err
	}
	//reddit returns arrays with mixed types (e.g. comments array will have a "more" element)
	//filter out anything that isn't type t1, which is a reddit comment
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

//FetchSubredditPosts fetches posts for a given subreddit e.g "jokes", "all"
//Reddit's api uses rate limiting, so any given request can have a 429 response
//if 429 is encountered, this function will sleep the thread for N milliseconds, then recursively try again
func FetchSubredditPosts(subreddit string, after string, before string, limit int) (resp RedditSubredditPostsResultContainer, nextAfter string, nextBefore string, err error) {
	//fmt.Println("\n ......fetching posts...... ")
	url := fmt.Sprintf("https://www.reddit.com/r/%s/top.json?limit=%s&after=%s&before=%s", subreddit, strconv.Itoa(limit), after, before)
	//make the request
	request, _ := http.NewRequest("GET", url, nil)
	request.Header.Add("Content-Type", "application/json")
	result := &RedditSubredditPostsResultContainer{}
	r, err := redditClient.Do(request)
	if err != nil {
		return *result, "", "", err
	}
	defer r.Body.Close()//close connection once func exits
	//reddit does rate limiting, so recursively call again in 1 second
	if r.StatusCode == 429 {
		//fmt.Println("done fetching. status code 429 (rate limiting by reddit) retrying shortly...: ", r.StatusCode)
		time.Sleep(rateLimitWaitMs * time.Millisecond)
		return FetchSubredditPosts(subreddit, after, before, limit)
	}
	//deserialize
	responseBody, _ := ioutil.ReadAll(r.Body)
	err = json.Unmarshal(responseBody, &result)
	return *result, result.Data.After, result.Data.Before, err
}
