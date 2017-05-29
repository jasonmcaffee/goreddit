package main

import (
	"fmt"
	"goreddit/goreddit"
	"flag"
	"os"
	"strconv"
)

func main(){
	subreddit, commentsPerPost := getApplicationFlags()

	fmt.Printf("fetching posts for: r/%s. comments per post: %s", subreddit, strconv.Itoa(commentsPerPost))
	goreddit.GoReddit(subreddit, commentsPerPost)
}

func getApplicationFlags() (subreddit string, commentsPerPost int){
	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	fs.StringVar(&subreddit, "subreddit", "all", "which subreddit to fetch posts from. e.g. 'all' for r/all")
	fs.IntVar(&commentsPerPost, "comments", 5, "how many comments to retrieve for each post")
	fs.Parse(os.Args[1:])
	return subreddit, commentsPerPost
}
