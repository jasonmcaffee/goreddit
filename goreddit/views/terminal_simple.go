package views

import (
	"fmt"
	"goreddit/goreddit/models"
	"strings"

	"github.com/fatih/color"
)

//CreatePostView returns an instance which implements IPostView (has Render func, etc)
func CreatePostView() IPostView {
	return &SimpleTerminalPostView{}
}

type (
	IPostView interface {
		Render(models.Post)
	}
)

type (
	SimpleTerminalPostView struct {
	}
)

//Render will display colored text for the post and its comments to the terminal window
func (s *SimpleTerminalPostView) Render(post models.Post) {
	//fmt.Println(color.RedString("########################################################################"))
	color.Cyan("------------------------------------------------------------------------------------------")
	c := color.New(color.BgCyan).Add(color.FgBlack).Add(color.Bold)
	c.Println(post.Title)

	c = color.New(color.FgCyan).Add(color.Bold)
	c.Println(post.Url)

	c = color.New(color.BgHiBlack).Add(color.Bold)
	fmt.Println(post.SelfText)
	s.renderComments(post)
}

//displays each comment for a given post
func (s *SimpleTerminalPostView) renderComments(post models.Post) {
	//fmt.Println("  -- Comments --  " + post.ID)
	for _, comment := range post.Comments {
		s.renderComment(comment)
	}
}

//displays comment for a given post
func (s *SimpleTerminalPostView) renderComment(comment string) {
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
