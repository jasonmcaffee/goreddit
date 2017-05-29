# goreddit
goreddit is a terminal application used to display reddit posts and comments in a terminal window.

## install
```
go get -u github.com/jasonmcaffee/goreddit
```

Ensure your $GOPATH environment variable is set and added to your $PATH
e.g. ~/.bash_profile
```
export GOPATH=/Users/you/Documents/dev/go
export GOROOT=/usr/local/Cellar/go/1.8/libexec
export PATH=$PATH:$GOPATH/bin
```
## run
```
goreddit
```

![Alt text](goreddit-terminal.png?raw=true "Optional Title")
## interactions
### urls
CMD + double-click on links inside the terminal, and they will be opened up in a new tab in your browser.

### flag options

```

```

#### -subreddit
name of the subreddit to retrieve posts from.
defaults to "all" for r/all

#### -comments
number of top comments to retrieve for each post
defaults to 5

#### -posts
number of posts to retrieve
defaults to 10