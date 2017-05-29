package goreddit

import (
	"goreddit/goreddit/reddit_client"
	"sync"

	"goreddit/goreddit/models"
	"goreddit/goreddit/views"
)

//GoReddit fetches posts and associated comments for a given subreddit, utilizing a IPostView to render them
//Channels are used to fetch posts in batches, fetch comments for posts, and rendering the post with comments.
func GoReddit(subreddit string) {
	postsChannel := make(chan ([]models.Post))
	postWithCommentsChannel := make(chan (models.Post))

	go fetchSubredditPostsUsingChannel(subreddit, 10, 10, "", "", postsChannel)
	go fetchCommentsForPostsReceivedOnChannel(5, postsChannel, postWithCommentsChannel)
	//block until all posts with comments have been received and displayed
	displayPostsWithCommentsReceivedOnChannel(postWithCommentsChannel)
}

//once a group of post has been retrieved on postsChannel, fetch comments for the post and send the post with comments
//to the passed in outChannel.
//WaitGroup is used so we can get comments for multiple posts at the same time.
func fetchCommentsForPostsReceivedOnChannel(commentLimit int, postsChannel chan ([]models.Post), outChannel chan (models.Post)) {
	var wg sync.WaitGroup
	for {
		select {
		case posts, ok := <-postsChannel:
			if !ok {
				goto DONERECEIVINGPOSTS
			}
			//asynchronously fetch comments for each post so we can attempt to get them all at once
			for _, post := range posts {
				wg.Add(1)
				go func(p models.Post) {
					p.Comments, _ = fetchCommentsForPost(p.ID, commentLimit)
					outChannel <- p
					wg.Done()
				}(post)
			}

		}
	}
DONERECEIVINGPOSTS:
	//fmt.Println("fetch comments done receiving posts")
	wg.Wait() //wait for all async calls to get comments for posts completes
	//fmt.Println("done receiving comments for all posts ")
	close(outChannel)
}

//wrapper for fetchSubredditPosts which utilizes recursion and channels to fetch posts in batches
//fetched posts will be sent to the passed in outChannel
func fetchSubredditPostsUsingChannel(subreddit string, limit int, remaining int, after string, before string, outChannel chan ([]models.Post)) {
	//fmt.Println("fetch subreddit posts using channel. limit:%s  remaining:%s", limit, remaining)
	posts, nextAfter, nextBefore, err := fetchSubredditPosts(subreddit, after, before, limit)
	if err != nil {
		close(outChannel)
		return
	}
	//send posts
	outChannel <- posts
	//determine if recursion is needed in order to fetch more posts
	remaining = remaining - len(posts)
	//fetch more if needed
	if remaining > 0 {
		fetchSubredditPostsUsingChannel(subreddit, limit, remaining, nextAfter, nextBefore, outChannel)
	} else {
		//fmt.Println("fetch subreddit posts closing channel")
		close(outChannel)
	}
}

//calls the reddit client to fetch posts, then maps to the Post model (reddits api models are too abstracted)
func fetchSubredditPosts(subreddit string, after string, before string, limit int) (posts []models.Post, nextAfter string, nextBefore string, err error) {
	postsResultsContainer, nAfter, nBefore, err := reddit_client.FetchSubredditPosts(subreddit, after, before, limit)
	posts = []models.Post{}
	if err == nil {
		posts = mapRedditPostsToPostModel(postsResultsContainer)
	}
	return posts, nAfter, nBefore, err
}

//calls the reddit client to fetch comments, then maps to a simpler model to represent a comment
func fetchCommentsForPost(postID string, limit int) ([]string, error) {
	redditCommentContainers, err := reddit_client.FetchCommentsForPost(postID, limit)
	commentStrings := []string{}
	if err == nil {
		commentStrings = mapRedditCommentsToCommentStrings(redditCommentContainers)
	}
	return commentStrings, err
}

//maps data from the reddit api model to a simpler model representation of comments
func mapRedditCommentsToCommentStrings(redditCommentContainers []reddit_client.RedditCommentContainer) []string {
	result := []string{}
	for _, redditCommentContainer := range redditCommentContainers {
		comment := redditCommentContainer.Data.Comment
		result = append(result, comment)
	}
	return result
}

//maps data from the reddit api model to a simpler model representation of posts
func mapRedditPostsToPostModel(redditPosts reddit_client.RedditSubredditPostsResultContainer) []models.Post {
	posts := []models.Post{}
	for _, redditPostContainer := range redditPosts.Data.Children {
		redditPost := redditPostContainer.Data
		post := models.Post{
			ID:       redditPost.ID,
			Title:    redditPost.Title,
			Url:      redditPost.Url,
			SelfText: redditPost.SelfText,
		}
		posts = append(posts, post)
	}
	return posts
}

//displays posts received on the passed in postChannel param
func displayPostsWithCommentsReceivedOnChannel(postsChannel chan (models.Post)) {
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

//creates a view for a post and renders it.
func displayPost(post models.Post) {
	views.CreatePostView().Render(post)
}
