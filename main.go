package main

import (
	"fmt"
	"goreddit/goreddit"
	"flag"
	"os"
)

func main(){
	subreddit := getApplicationFlags()

	fmt.Println("fetching posts for ", subreddit)
	goreddit.GoReddit(subreddit)
}

func getApplicationFlags() (subreddit string){
	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	fs.StringVar(&subreddit, "subreddit", "all", "which subreddit to fetch posts from. e.g. 'all' for r/all")
	fs.Parse(os.Args[1:])
	return subreddit
}
