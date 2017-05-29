package goreddit

import (
	"fmt"
	"goreddit/goreddit/reddit_client"
	"strings"
	"sync"

	"github.com/fatih/color"
)

//view model
type (
	Post struct {
		ID       string
		Url      string
		Title    string
		Comments []string
	}
)

//We could have fetchSubredditPosts fetch comments, and make the code a bit easier, but we are intentionally
//showing sophisticated use of channels
//fetch posts in batches to demonstrate channel usage
//fetch comments for each post received
func GoReddit() {
	postsChannel := make(chan ([]Post))
	postWithCommentsChannel := make(chan (Post))

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		fetchSubredditPostsUsingChannel(5, 25, "", "", postsChannel)
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		fetchCommentsForPostsReceivedOnChannel(5, postsChannel, postWithCommentsChannel)
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		displayPostsWithCommentsReceivedOnChannel(postWithCommentsChannel)
		wg.Done()
	}()

	wg.Wait()
}

func fetchCommentsForPostsReceivedOnChannel(commentLimit int, postsChannel chan ([]Post), outChannel chan (Post)) {
	var wg sync.WaitGroup
	for {
		select {
		case posts, ok := <-postsChannel:
			if !ok {
				goto DONERECEIVINGPOSTS
			}
			//asynchronously fetch comments for each post
			for _, post := range posts {
				wg.Add(1)
				go func(p Post) {
					p.Comments, _ = fetchCommentsForPost(p.ID, commentLimit)
					outChannel <- p
					wg.Done()
				}(post)
			}

		}
	}
DONERECEIVINGPOSTS:
	//fmt.Println("fetch comments done receiving posts")
	wg.Wait()
	//fmt.Println("done receiving comments for all posts ")
	close(outChannel)
}

func fetchSubredditPostsUsingChannel(limit int, remaining int, after string, before string, outChannel chan ([]Post)) {
	//fmt.Println("fetch subreddit posts using channel. limit:%s  remaining:%s", limit, remaining)
	posts, nextAfter, nextBefore, err := fetchSubredditPosts(after, before, limit)
	if err != nil {
		close(outChannel)
		return
	}

	//send posts
	outChannel <- posts

	remaining = remaining - len(posts)
	//fetch more if needed
	if remaining > 0 {
		fetchSubredditPostsUsingChannel(limit, remaining, nextAfter, nextBefore, outChannel)
	} else {
		//fmt.Println("fetch subreddit posts closing channel")
		close(outChannel)
	}
}

func fetchSubredditPosts(after string, before string, limit int) (posts []Post, nextAfter string, nextBefore string, err error) {
	postsResultsContainer, nAfter, nBefore, err := reddit_client.FetchSubredditPosts(after, before, limit)
	posts = []Post{}
	if err != nil {
		posts = mapRedditPostsToOurModel(postsResultsContainer)
	}
	return posts, nAfter, nBefore, err
}

func fetchCommentsForPost(postID string, limit int) ([]string, error) {
	redditCommentContainers, err := reddit_client.FetchCommentsForPost(postID, limit)
	commentStrings := []string{}
	if err != nil {
		commentStrings = mapRedditCommentsToCommentStrings(redditCommentContainers)
	}
	return commentStrings, err
}

func mapRedditCommentsToCommentStrings(redditCommentContainers []reddit_client.RedditCommentContainer) []string {
	result := []string{}
	for _, redditCommentContainer := range redditCommentContainers {
		comment := redditCommentContainer.Data.Comment
		result = append(result, comment)
	}
	return result
}

func mapRedditPostsToOurModel(redditPosts reddit_client.RedditSubredditPostsResultContainer) []Post {
	posts := []Post{}
	for _, redditPostContainer := range redditPosts.Data.Children {
		redditPost := redditPostContainer.Data
		post := Post{
			ID:    redditPost.ID,
			Title: redditPost.Title,
			Url:   redditPost.Url,
		}
		posts = append(posts, post)
	}
	return posts
}

func displayPostsWithCommentsReceivedOnChannel(postsChannel chan (Post)) {
	for {
		select {
		case post, ok := <-postsChannel:
			if !ok {
				goto DONERECEIVINGPOSTS
			}
			displayPost(post)
		}
	}
DONERECEIVINGPOSTS:
	//fmt.Println("display posts with comments done receiving")
}

func displayPosts(posts []Post) {
	//iterate over each child
	for _, post := range posts {
		displayPost(post)
	}
}

func displayPost(post Post) {
	fmt.Println(color.RedString("########################################################################"))
	color.Cyan("------------------------------------------------------------------------------------------")
	fmt.Println("### ", post.Title)
	fmt.Println("", post.Url)
	displayComments(post)
}

func displayComments(post Post) {
	//fmt.Println("  -- Comments --  " + post.ID)
	for _, comment := range post.Comments {
		displayComment(comment)
	}
}

func displayComment(comment string) {
	commentLines := strings.Split(comment, "\n")
	fmt.Print("   -  ") //indicate comment and indent
	for i, commentLine := range commentLines {
		if i == 0 {
			fmt.Println(commentLine)
		} else {
			fmt.Println("      " + commentLine)
		}
	}
}

//posts, err := fetchSubredditPosts("", "", 200)
//if err != nil {
//fmt.Println("error fetching: ", err)
//return
//}
//displayRedditSubredditPosts(posts)

//
//func fetchSubredditPosts(after string, before string, total int) (RedditSubredditPostsResult, error) {
//	limit := 50
//	url := getUrl(after, before, limit) //"https://www.reddit.com/r/all/top.json?limit=100&count=100&after=100" //t3_XXXXX
//	fmt.Println("fetching posts...", url)
//	//url := "https://server2.kproxy.com/servlet/redirect.srv/sruj/scsoprf/smno/p2/r/all/top.json?count=N&after=t3_XXXXX"
//	redditClient := &http.Client{Timeout: 60 * time.Second}
//	request, _ := http.NewRequest("GET", url, nil)
//	request.Header.Add("Content-Type", "application/json")
//
//	result := &RedditSubredditPostsResult{}
//	r, err := redditClient.Do(request)
//	if err != nil {
//		return *result, err
//	}
//	defer r.Body.Close()
//	responseBody, _ := ioutil.ReadAll(r.Body)
//	err = json.Unmarshal(responseBody, &result)
//
//	if r.StatusCode == 429 {
//		fmt.Println("done fetching. status code 429 (rate limiting by reddit) retrying shortly...: ", r.StatusCode)
//		time.Sleep(1 * time.Millisecond)
//		return fetchSubredditPosts(after, before, total)
//	}
//
//	remainingPostsToFetch := total - limit
//	if remainingPostsToFetch > 0 {
//		nextAfter := result.Data.After
//		nextBefore := result.Data.Before
//		nextResult, err := fetchSubredditPosts(nextAfter, nextBefore, remainingPostsToFetch)
//		if err == nil {
//			result.Data.Children = append(result.Data.Children, nextResult.Data.Children...)
//			fmt.Println("============== NEW TOTAL =============  ", len(result.Data.Children))
//		}
//	}
//
//	return *result, err
//}

//jsonStr := string(responseBody[:])
////rawJson, err := json.Marshal(&responseBody)
////jsonStr := string(rawJson)
//fmt.Println(jsonStr)

//
//func commentParse(){
//	//commentResponse := `[{"kind": "Listing", "data": {"children":[ {"kind":"123", "data":{"body":"hello"} } ]} }]`
//	commentResponse := ``
//	rawComments := json.RawMessage(commentResponse)
//	commentBytes, err := json.Marshal(rawComments)
//	fmt.Println(err)
//	results := []RedditCommentResponse{}
//	err = json.Unmarshal(commentBytes, &results)
//	fmt.Println(err)
//	fmt.Printf("%v",results)
//	fmt.Println(results[0].Data.Children[0].Data.Comment)
//}
