/**
 *   ____     _____          _ _       _
 *  / ___| __|_   _|_      _(_) |_ ___| |__
 * | |  _ / _ \| | \ \ /\ / / | __/ __| '_ \
 * | |_| | (_) | |  \ V  V /| | || (__| | | |
 *  \____|\___/|_|   \_/\_/ |_|\__\___|_| |_|
 *
 *    Abuse me with your shell scripts!
 */

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/deluxo/gotwitchlib"
	"github.com/fatih/color"
	"gopkg.in/alecthomas/kingpin.v2"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
)

// User Model
type User struct {
	Username   string
	OauthToken string
}

// Player Model
type Player struct {
	Name string
}

// Options for Settings
type Options struct {
	Game    bool
	Status  bool
	Padding int
}

// Settings that ar saved to home/.config/gotwitch
type Settings struct {
	User    User
	Player  Player
	Options Options
}

// ErrNoChannelProvided error
const ErrNoChannelProvided = "No channel name provided. Please write channel name in command arguments..."

var (
	s              = getSettings()
	twitchClientID = "ctf0u38gzxl1emqdrsp17y0e20o1ajh"
	twitchRedirURL = "https://deluxo.github.io/gotwitch/"
	wr             = bufio.NewWriter(os.Stdout)
	usr, _         = user.Current()
	settingsPath   = usr.HomeDir + "/.config/gotwitch/config.json"
	app            = kingpin.New("gotwitch", "A command-line twitch.tv application written in Golang.")

	printColPadding = app.Flag("padding", "output column padding width").Default("25").Short('d').Int()
	streamer        = app.Command("streamer", "Do actions related to a streamer").Default()
	game            = app.Command("game", "Do actions related to a game")
	setup           = app.Command("setup", "setup procedure")

	gameTitle       = game.Arg("title", "game title a.k.a category").String()
	streamerChannel = streamer.Arg("channel", "channel of a streamer").HintAction(listChannels).String()

	streamerURL           = streamer.Flag("url", "print channels full URL for lazy people").Short('r').Bool()
	streamerWatch         = streamer.Flag("watch", "watch the stream through a given player").Short('w').Bool()
	streamerPlayer        = streamer.Flag("player", "player to use for watching a stream").Short('p').Default("mpv").String()
	streamerFollow        = streamer.Flag("follow", "follow the streamer").Short('f').Bool()
	streamerFollowNotify  = streamer.Flag("notify", "notify if the streamer comes online").Short('n').Bool()
	streamerUnfollow      = streamer.Flag("unfollow", "unfollow the streamer").Short('u').Bool()
	streamerSearch        = streamer.Flag("query", "query for streams with a game or channel name").Short('q').String()
	streamerList          = streamer.Flag("ls", "list the streamers").Short('l').Bool()
	streamerSubscribed    = streamer.Flag("subscribed", "filter out only subscribed streamers").Short('b').Bool()
	streamerThumbnail     = streamer.Flag("thumbnail", "include streamer thumbnail").Short('t').Bool()
	streamerIngludeGame   = streamer.Flag("game", "print the game a streamer is playing").Short('g').Bool()
	streamerIncludeStatus = streamer.Flag("status", "print the streamer's status").Short('s').Bool()
	streamerOffset        = streamer.Flag("offset", "streamer list view starting point").Default("0").Short('o').Int()
	streamerLimit         = streamer.Flag("limit", "streamer list view length").Default("10").Short('i').Int()

	gameOffset = game.Flag("offset", "game list view starting point").Default("0").Short('o').Int()
	gameLimit  = game.Flag("limit", "game list view length").Default("10").Short('i').Int()

	setupUser        = setup.Flag("username", "twitch.tv channel").String()
	setupAccessToken = setup.Flag("access-token", "a generated access token provided by twitch.tv").Default("generate").String()
	setupPlayer      = setup.Flag("player", "video player to use for stream watching by default").String()
)

func main() {
	kingpin.CommandLine.HelpFlag.Short('h')

	player := s.Player.Name
	if *streamerPlayer != "" {
		player = *streamerPlayer
	}

	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case streamer.FullCommand():
		switch {
		case *streamerWatch:
			if *streamerChannel == "" {
				fmt.Println(ErrNoChannelProvided)
			} else {
				exec.Command(
					player,
					twitch.TwitchURL+*streamerChannel,
				).Start()
			}
		case *streamerURL:
			if *streamerChannel != "" {
				fmt.Println(twitch.TwitchURL + *streamerChannel)
			} else {
				fmt.Println(ErrNoChannelProvided)
			}
		case *streamerFollow && !*streamerUnfollow:
			response := twitch.Follow(
				s.User.OauthToken,
				s.User.Username,
				*streamerChannel,
				*streamerFollowNotify,
			)
			printFollow(response)
		case *streamerUnfollow && !*streamerFollow:
			response := twitch.Unfollow(
				s.User.OauthToken,
				s.User.Username,
				*streamerChannel,
			)
			printFollow(response)
		case *streamerSearch != "":
			for _, v := range twitch.SearchStreams(s.User.OauthToken, streamerLimit, streamerOffset, streamerSearch).Streams {
				printStream(v.Channel, streamerIncludeStatus, streamerIngludeGame, streamerThumbnail)
			}
		case *streamerList:
			if *streamerSubscribed {
				for _, v := range twitch.GetLiveSubs(s.User.OauthToken).Streams {
					printStream(v.Channel, streamerIncludeStatus, streamerIngludeGame, streamerThumbnail)
				}
			} else {
				for _, v := range twitch.GetStreams(s.User.OauthToken, "", "", 0, 0).Streams {
					printStream(v.Channel, streamerIncludeStatus, streamerIngludeGame, streamerThumbnail)
				}
			}
		}

	case game.FullCommand():
		games := twitch.GetTopGames(s.User.OauthToken, gameLimit, gameOffset)
		for _, v := range games.Top {
			printGame(v.Game)
		}

	case setup.FullCommand():
		if *setupAccessToken == "generate" {
			url := "https://api.twitch.tv/kraken/oauth2/authorize?response_type=token&client_id=" + twitchClientID + "&redirect_uri=" + twitchRedirURL + "&scope=user_read+user_blocks_edit+user_blocks_read+user_follows_edit+channel_read+channel_editor+channel_commercial+channel_stream+channel_subscriptions+user_subscriptions+channel_check_subscription+chat_login+channel_feed_read+channel_feed_edit"
			exec.Command("xdg-open", url).Start()
		} else if *setupUser != "" && *setupPlayer != "" && *setupAccessToken != "" {
			setSettings(*setupUser, *setupAccessToken, *setupPlayer, *printColPadding)
		} else {
			fmt.Println("Not generating the access token, nor creating the config file. Nothing to do...")
		}
	}
}

func listChannels() []string {
	s := getSettings()

	subStreams := twitch.GetLiveSubs(
		s.User.OauthToken).Streams

	popularStreams := twitch.GetStreams(s.User.OauthToken, "", "", 0, 0).Streams
	var list []string
	for _, v := range subStreams {
		list = append(list, v.Channel.Name)
	}
	for _, v := range popularStreams {
		fmt.Fprintln(wr, v.Channel.Name)
		list = append(list, v.Channel.Name)
	}

	return list
}

func getSettings() Settings {
	var settings Settings
	settingsFile, err := os.Open(settingsPath)
	if err == nil {
		fileStream, _ := ioutil.ReadAll(settingsFile)
		json.Unmarshal(fileStream, &settings)
	}
	return settings
}

func printGame(s twitch.Game) {
	game := color.New(color.Bold, color.FgHiRed).SprintFunc()
	lineColored := game(s.Name)
	fmt.Fprintln(wr, lineColored)

	defer wr.Flush()
	defer color.Unset()
}

func printStream(s twitch.Channel, showFlag *bool, gameFlag *bool, thumbnailFlag *bool) {
	nick := color.New(color.FgHiBlue).SprintFunc()
	status := color.New(color.FgHiWhite).SprintFunc()
	game := color.New(color.Bold, color.FgHiRed).SprintFunc()
	thumb := color.New(color.Bold, color.FgBlack).SprintFunc()
	lineColored := nick(s.Name)
	if *showFlag == true {
		sp := *printColPadding - len(lineColored)
		if sp < *printColPadding {
			for i := 0; i < sp; i++ {
				lineColored += " "
			}
		}
		lineColored += " " + status(s.Status)
	}
	if *gameFlag == true {
		sp := *printColPadding - len(lineColored)
		if sp < *printColPadding {
			for i := 0; i < sp; i++ {
				lineColored += " "
			}
		}
		lineColored += " " + game(s.Game)
	}
	if *thumbnailFlag == true {
		sp := *printColPadding - len(lineColored)
		if sp < *printColPadding {
			for i := 0; i < sp; i++ {
				lineColored += " "
			}
		}
		lineColored += " " + thumb("("+s.Logo+")")
	}
	fmt.Fprintln(wr, lineColored)

	defer wr.Flush()
	defer color.Unset()
}

func printFollow(s twitch.FollowChannel) {
	fmt.Fprintln(wr, "ok")
	defer wr.Flush()
}

func replaceAtIndex(in string, r rune, i int) string {
	out := []rune(in)
	out[i] = r
	return string(out)
}

func limitStringLength(line string, maxLineLength int) string {
	if len(line) > maxLineLength {
		line = line[:maxLineLength]
		for i := 1; i < 4; i++ {
			line = replaceAtIndex(line, '.', len(line)-i)
		}
	}
	return line
}

func setSettings(twitchUser, twitchOauthToken, playerName string, padding int) {
	settings := Settings{
		User: User{
			Username:   twitchUser,
			OauthToken: twitchOauthToken,
		},
		Player: Player{
			Name: playerName,
		},
		Options: Options{
			Game:    false,
			Status:  false,
			Padding: padding,
		},
	}
	settingsJSON, _ := json.MarshalIndent(settings, "", "\t")
	fmt.Printf(
		"Your settings are located at:\n%s\n\nand look like so:\n%s\n",
		settingsPath,
		settingsJSON,
	)
	os.MkdirAll(filepath.Dir(settingsPath), 0775)
	ioutil.WriteFile(settingsPath, settingsJSON, 776)
}
