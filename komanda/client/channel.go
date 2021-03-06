package client

import (
	"fmt"
	"sort"
	"sync"

	"github.com/hectane/go-nonblockingchan"
	"github.com/jroimartin/gocui"
	"github.com/mephux/komanda-cli/komanda/color"
	"github.com/mephux/komanda-cli/komanda/config"
)

// RenderHandlerFunc type for Exec callbacks
type RenderHandlerFunc func(*Channel, *gocui.View) error

// User struct
type User struct {
	Nick  string
	Mode  string
	Color int
}

// String converts the user struct to a string
func (u *User) String(c bool) string {
	if u == nil {
		return ""
	}

	if c {
		return color.Stringf(u.Color, "%s%s", u.Mode, u.Nick)
	}

	return fmt.Sprintf("%s%s", u.Mode, u.Nick)
}

// NickSorter cast to an array of user pointers
type NickSorter []*User

// Len returns the list length
func (a NickSorter) Len() int { return len(a) }

// Swap moves the position of two items in the list
func (a NickSorter) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

// Less checks if i is less than j to change position
func (a NickSorter) Less(i, j int) bool { return a[i].Nick < a[j].Nick }

// Channel struct
type Channel struct {
	Status        bool
	Ready         bool
	Unread        bool
	Highlight     bool
	Name          string
	Server        *Server
	MaxX          int
	MaxY          int
	RenderHandler RenderHandlerFunc
	Topic         string
	TopicSetBy    string
	Users         []*User
	NickListReady bool
	Loading       *nbc.NonBlockingChan
	Private       bool

	mu sync.Mutex
}

// FindUser returns a pointer for a given user or nil
func (channel *Channel) FindUser(nick string) *User {
	for _, u := range channel.Users {
		if u.Nick == nick {
			return u
		}
	}

	return nil
}

// View returns the channel view
func (channel *Channel) View() (*gocui.View, error) {
	return channel.Server.Gui.View(channel.Name)
}

// Current returns true or false if the current channel is this channel
func (channel *Channel) Current() bool {
	if channel.Server.CurrentChannel == channel.Name {
		return true
	}

	return false
}

// Update will render the current channel again
func (channel *Channel) Update() (*gocui.View, error) {
	channel.MaxX, channel.MaxY = channel.Server.Gui.Size()

	return channel.Server.Gui.SetView(channel.Name,
		-1, -1, channel.MaxX, channel.MaxY-4)

}

// NickListString will output the channel users in a pretty format
func (channel *Channel) NickListString(v *gocui.View, c bool) {
	sort.Sort(NickSorter(channel.Users))

	fmt.Fprintf(v, "\n%s", color.String(config.C.Color.Green, "== NICK LIST START\n"))

	for i, u := range channel.Users {
		if i == len(channel.Users)-1 {
			fmt.Fprintf(v, "%s", u.String(c))
		} else {
			fmt.Fprintf(v, "%s, ", u.String(c))
		}
	}

	fmt.Fprintf(v, "\n%s", color.String(config.C.Color.Green, "== NICK LIST END\n\n"))
}

// NickMetricsString will output channel metrics in a pretty format
// 09:41 * Irssi: #google-containers: Total of 213 nicks [0 ops, 0 halfops, 0 voices, 213 normal]
func (channel *Channel) NickMetricsString(view *gocui.View) {
	var op, hop, v, n int

	for _, u := range channel.Users {
		switch u.Mode {
		case "@":
			op++
		case "%":
			hop++
		case "+":
			v++
		default:
			n++
		}
	}

	fmt.Fprintf(view, "%s Komanda: %s: Total of %d nicks [%d ops, %d halfops, %d voices, %d normal]\n\n",
		color.String(config.C.Color.Green, "**"), channel.Name, len(channel.Users), op, hop, v, n)
}

// RemoveNick from channel list
func (channel *Channel) RemoveNick(nick string) {
	for i, user := range channel.Users {
		if user.Nick == nick {
			channel.mu.Lock()
			defer channel.mu.Unlock()

			channel.Users = append(channel.Users[:i], channel.Users[i+1:]...)
		}
	}
}

// AddNick to channel list
func (channel *Channel) AddNick(nick string) {

	if u := channel.FindUser(nick); u == nil {
		channel.mu.Lock()
		defer channel.mu.Unlock()

		user := &User{
			Nick:  nick,
			Color: color.Random(22, 231),
		}

		channel.Users = append(channel.Users, user)
	}
}

// Render the current channel
func (channel *Channel) Render(update bool) error {

	view, err := channel.Server.Gui.SetView(channel.Name,
		-1, -1, channel.MaxX, channel.MaxY-2)

	if err != gocui.ErrUnknownView {
		return err
	}

	if channel.Name != StatusChannel {
		view.Autoscroll = true
		view.Wrap = true
		// view.Highlight = true
		view.Frame = false

		view.FgColor = gocui.ColorWhite
		// view.BgColor = gocui.ColorDefault
		view.BgColor = gocui.ColorWhite
		// view.BgColor = gocui.Attribute(0)

		if !channel.Private {
			fmt.Fprint(view, "\n\n\n")
		} else {
			channel.Topic = fmt.Sprintf("Private Chat: %s", channel.Name)
			fmt.Fprint(view, "\n\n")
		}
	}

	view.Wrap = true

	if !update {
		if err := channel.RenderHandler(channel, view); err != nil {
			return err
		}
	}

	if channel.Private {
		channel.Server.Gui.SetViewOnTop(channel.Server.CurrentChannel)
	}

	return nil
}
