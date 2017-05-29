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
		SelfText string
		Comments []string
	}
)

//We could have fetchSubredditPosts fetch comments, and make the code a bit easier, but we are intentionally
//showing sophisticated use of channels
//fetch posts in batches to demonstrate channel usage
//fetch comments for each post received
func GoReddit(subreddit string) {
	postsChannel := make(chan ([]Post))
	postWithCommentsChannel := make(chan (Post))

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		fetchSubredditPostsUsingChannel(subreddit, 10, 10, "", "", postsChannel)
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

func fetchSubredditPostsUsingChannel(subreddit string, limit int, remaining int, after string, before string, outChannel chan ([]Post)) {
	//fmt.Println("fetch subreddit posts using channel. limit:%s  remaining:%s", limit, remaining)
	posts, nextAfter, nextBefore, err := fetchSubredditPosts(subreddit, after, before, limit)
	if err != nil {
		close(outChannel)
		return
	}

	//send posts
	outChannel <- posts

	remaining = remaining - len(posts)
	//fetch more if needed
	if remaining > 0 {
		fetchSubredditPostsUsingChannel(subreddit, limit, remaining, nextAfter, nextBefore, outChannel)
	} else {
		//fmt.Println("fetch subreddit posts closing channel")
		close(outChannel)
	}
}

func fetchSubredditPosts(subreddit string, after string, before string, limit int) (posts []Post, nextAfter string, nextBefore string, err error) {
	postsResultsContainer, nAfter, nBefore, err := reddit_client.FetchSubredditPosts(subreddit, after, before, limit)
	posts = []Post{}
	if err == nil {
		posts = mapRedditPostsToOurModel(postsResultsContainer)
	}
	return posts, nAfter, nBefore, err
}

func fetchCommentsForPost(postID string, limit int) ([]string, error) {
	redditCommentContainers, err := reddit_client.FetchCommentsForPost(postID, limit)
	commentStrings := []string{}
	if err == nil {
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
			SelfText: redditPost.SelfText,
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
	//fmt.Println(color.RedString("########################################################################"))
	color.Cyan("------------------------------------------------------------------------------------------")
	fmt.Println("### ", post.Title)
	fmt.Println("", post.Url)
	fmt.Println("", post.SelfText)
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