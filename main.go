package main

import (
	"fmt"
	"goreddit/goreddit"
)

func main(){
	fmt.Println("fetching posts")
	goreddit.GoReddit()
}
