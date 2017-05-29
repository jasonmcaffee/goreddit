package views

import (
	"fmt"
	"goreddit/goreddit/models"
	"strings"

	"github.com/fatih/color"
)

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

func (s *SimpleTerminalPostView) Render(post models.Post) {
	//fmt.Println(color.RedString("########################################################################"))
	color.Cyan("------------------------------------------------------------------------------------------")
	c := color.New(color.BgCyan).Add(color.FgBlack).Add(color.Bold)
	c.Println(post.Title)

	c = color.New(color.FgCyan).Add(color.Bold)
	c.Println(post.Url)

	c = color.New(color.BgHiBlack).Add(color.Bold)
	fmt.Println(post.SelfText)
	s.displayComments(post)
}

func (s *SimpleTerminalPostView) displayComments(post models.Post) {
	//fmt.Println("  -- Comments --  " + post.ID)
	for _, comment := range post.Comments {
		s.displayComment(comment)
	}
}

func (s *SimpleTerminalPostView) displayComment(comment string) {
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
