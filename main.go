package main

import (
	"fmt"
	"goreddit/goreddit"
	"flag"
	"os"
	"strconv"
)

func main(){
	subreddit, commentsPerPost, postsCount := getApplicationFlags()

	fmt.Printf("fetching posts for: r/%s. posts to retrieve: %s  comments per post: %s \n", subreddit, strconv.Itoa(postsCount), strconv.Itoa(commentsPerPost))
	goreddit.GoReddit(subreddit, commentsPerPost, postsCount)
}

func getApplicationFlags() (subreddit string, commentsPerPost int, postsCount int){
	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	fs.StringVar(&subreddit, "subreddit", "all", "which subreddit to fetch posts from. e.g. 'all' for r/all")
	fs.IntVar(&commentsPerPost, "comments", 5, "how many comments to retrieve for each post")
	fs.IntVar(&postsCount, "posts", 10, "how many posts to retrieve")
	fs.Parse(os.Args[1:])
	return subreddit, commentsPerPost, postsCount
}
