package views

import (
	"fmt"
	"goreddit/goreddit/models"
	"strings"
	"syscall"
	"unsafe"
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

//38 normal
//39 blink no background color
//40 blink black background
//41
//42 blink background color
//48 background
const (
	ColorPurple string = "20"
	ColorGreen string = "40"
	ColorNone string = ""
	ColorLightBlue string = "81"
	ColorLightGrey string = "252"
	ColorDarkGrey string = "235"
)

func minWidth(text string, min int, padChar string) string {
	if len(text) > min {
		return text
	}
	neededPadding := min - len(text)
	result := text
	for i:=0; i < neededPadding; i++{
		result += padChar
	}
	return result
}

//colorFormat
func colorize(text string, fgColor string, bgColor string) string{
	fgColored := fmt.Sprintf("\x1b[38;5;%vm%s\x1b[0m", fgColor, text)
	if bgColor == ColorNone {
		return fgColored
	}
	bgColored := fmt.Sprintf("\x1b[48;5;%vm%s\x1b[0m", bgColor, fgColored)
	return bgColored

}

//http://misc.flogisoft.com/bash/tip_colors_and_formatting
//Render will display colored text for the post and its comments to the terminal window
func (s *SimpleTerminalPostView) Render(post models.Post) {
	minimumWidth := 100
	//hr
	fmt.Println(colorize(minWidth("", minimumWidth, "\u005F"), ColorLightBlue, ColorNone))
	//title
	fmt.Println(colorize(minWidth(post.Title, minimumWidth, " "), ColorDarkGrey, ColorLightGrey))
	//url
	fmt.Println(colorize(minWidth(post.Url, minimumWidth, " "), ColorLightBlue, ColorNone))
	//self text
	if len(post.SelfText) > 0{
		fmt.Println(minWidth(post.SelfText, minimumWidth, " "))
	}

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

type winsize struct {
	Row    uint16
	Col    uint16
	Xpixel uint16
	Ypixel uint16
}

func getWidth() uint {
	ws := &winsize{}
	retCode, _, errno := syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(syscall.Stdin),
		uintptr(syscall.TIOCGWINSZ),
		uintptr(unsafe.Pointer(ws)))

	if int(retCode) == -1 {
		panic(errno)
	}
	return uint(ws.Col)
}
