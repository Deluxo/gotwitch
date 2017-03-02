package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/deluxo/gotwitchlib"
	"github.com/fatih/color"
	"gopkg.in/alecthomas/kingpin.v2"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
)

type User struct {
	Username   string
	OauthToken string
}
type Player struct {
	Name string
}
type Options struct {
	Game    bool
	Status  bool
	Padding int
}
type Settings struct {
	User    User
	Player  Player
	Options Options
}

var (
	twitchClientId = "ctf0u38gzxl1emqdrsp17y0e20o1ajh"
	twitchRedirUrl = "https://deluxo.github.io/gotwitch/"
	strmLen        = 20
	wr             = bufio.NewWriter(os.Stdout)
	usr, _         = user.Current()
	settingsPath   = usr.HomeDir + "/.config/gotwitch/config.json"

	app = kingpin.New("gotwitch", "A command-line twitch.tv application written in Golang.")

	// streams
	streams              = app.Command("streams", "Look up online streams.")
	streamsGame          = streams.Flag("game", "Game title, like dota2").Short('g').String()
	streamsType          = streams.Flag("type", "Stream type, e.g. live, all, or playlist").Short('t').String()
	streamsLimit         = streams.Flag("limit", "Number of streams to print.").Short('l').Int()
	streamsOffset        = streams.Flag("offset", "Pagination offset for given limit.").Short('o').Int()
	streamsSubscribtions = streams.Flag("subscribtions", "Filter only streams that are subscribed.").Short('b').Bool()
	//streamsNotify        = streams.Flag("notify", "Print the output through the notification instead of std out.").Short('n').Bool()
	streamsPrintStatus = streams.Flag("status", "Include stream status into output.").Short('u').Bool()
	streamsPrintGame   = streams.Flag("print-game", "Include streamded game title into output.").Short('a').Bool()

	// follow
	follow             = app.Command("follow", "Follow the streamer.")
	followStreamer     = follow.Flag("streamer", "Streamer name.").HintAction(listChannels).Short('s').String()
	followNotification = follow.Flag("notify", "Get notified when the streamer comes online.").Short('n').Bool()

	// follow
	unfollow             = app.Command("unfollow", "Unfollow the streamer.")
	unfollowStreamer     = unfollow.Flag("streamer", "Streamer name.").HintAction(listChannels).Short('s').String()

	// watch
	watch    = app.Command("watch", "Watch the stream.").Default()
	streamer = watch.Flag("streamer", "Streamer name.").HintAction(listChannels).Short('s').String()

	// games
	topGames       = app.Command("games", "Get the list of currently most played games.")
	limit          = topGames.Flag("limit", "Wanted list size of a request.").Int()
	offset         = topGames.Flag("offset", "Pagination offset of a page (default page size is 10).").Int()
	topGamesNotify = topGames.Flag("notify", "Print the output through the notification instead of std out.").Short('n').Bool()

	// setup
	setup            = app.Command("setup", "create a config file with required values")
	twitchUser       = setup.Flag("username", "Twitch.tv username.").Short('u').String()
	setupAccessToken = setup.Flag("auth", "generate oAuthToken (found in URL after successful login).").Short('a').Bool()
	twitchOauthToken = setup.Flag("oauth", "Twitch.tv oAuthToken (must be generated first).").Short('o').String()
	playerName       = setup.Flag("player", "Player command to be used, like: mpv or vlc.").Short('p').String()
)

func main() {
	kingpin.CommandLine.HelpFlag.Short('h')
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {

	case streams.FullCommand():
		s := getSettings()
		strmLen = s.Options.Padding
		if s.Options.Game == true {
			*streamsPrintGame = true
		}
		if s.Options.Status == true {
			*streamsPrintStatus = true
		}
		if *streamsSubscribtions == true {
			for _, v := range twitch.GetLiveSubs(getSettings().User.OauthToken).Streams {
				printStream(v.Channel, streamsPrintStatus, streamsPrintGame)
			}
		} else {
			for _, v := range twitch.GetStreams(getSettings().User.OauthToken, *streamsGame, *streamsType, *streamsLimit, *streamsOffset).Streams {
				printStream(v.Channel, streamsPrintStatus, streamsPrintGame)
			}
		}

	case follow.FullCommand():
		s := getSettings()
		response := twitch.Follow(s.User.OauthToken, s.User.Username, *followStreamer, *followNotification)
		printFollow(response)

	case unfollow.FullCommand():
		s := getSettings()
		response := twitch.Unfollow(s.User.OauthToken, s.User.Username, *unfollowStreamer,)
		printFollow(response)

	case topGames.FullCommand():
		games := twitch.GetTopGames(getSettings().User.OauthToken, limit, offset)
		if *topGamesNotify == true {
			var notification string
			for k, v := range games.Top {
				notification += strconv.Itoa(k+1) + ". " + v.Game.Name + "\n"
			}
			printNotification(notification)
		} else {
			for _, v := range games.Top {
				printGame(v.Game)
			}
		}

	case watch.FullCommand():
		s := getSettings()
		exec.Command(s.Player.Name, twitch.TwitchURL+*streamer).Start()

	case setup.FullCommand():
		if *setupAccessToken {
			url := "https://api.twitch.tv/kraken/oauth2/authorize?response_type=token&client_id=" + twitchClientId + "&redirect_uri=" + twitchRedirUrl + "&scope=user_read+user_blocks_edit+user_blocks_read+user_follows_edit+channel_read+channel_editor+channel_commercial+channel_stream+channel_subscriptions+user_subscriptions+channel_check_subscription+chat_login+channel_feed_read+channel_feed_edit"
			exec.Command("xdg-open", url).Start()
		} else if *twitchUser != "" && *playerName != "" && *twitchOauthToken != "" {
			setSettings(*twitchUser, *twitchOauthToken, *playerName)
		} else {
			fmt.Println("Not generating the access token, nor creating the config file. Nothing to do...")
		}
	}
}

func listChannels() []string {
	s := getSettings()

	subStreams := twitch.GetLiveSubs(
		s.User.OauthToken).Streams

	popularStreams := twitch.GetStreams(
		getSettings().User.OauthToken,
		*streamsGame,
		*streamsType,
		*streamsLimit,
		*streamsOffset,
	).Streams
	list := make([]string, 0)
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
	if err != nil {
		fmt.Println("No config file was found. run gotwitch setup --help")
		os.Exit(0)
	}
	fileStream, _ := ioutil.ReadAll(settingsFile)
	json.Unmarshal(fileStream, &settings)
	return settings
}

func printGame(s twitch.Game) {
	game := color.New(color.Bold, color.FgHiRed).SprintFunc()
	lineColored := game(s.Name)
	fmt.Fprintln(wr, lineColored)

	defer wr.Flush()
	defer color.Unset()
}

func printStream(s twitch.Channel, showFlag *bool, gameFlag *bool) {
	nick := color.New(color.FgHiBlue).SprintFunc()
	status := color.New(color.FgHiWhite).SprintFunc()
	game := color.New(color.Bold, color.FgHiRed).SprintFunc()
	lineColored := nick(s.Name)
	if *showFlag == true {
		sp := strmLen - len(lineColored)
		if sp < strmLen {
			for i := 0; i < sp; i++ {
				lineColored += " "
			}
		}
		lineColored += " " + status(s.Status)
	}
	if *gameFlag == true {
		sp := strmLen - len(lineColored)
		if sp < strmLen {
			for i := 0; i < sp; i++ {
				lineColored += " "
			}
		}
		lineColored += " " + game(s.Game)
	}
	fmt.Fprintln(wr, lineColored)

	defer wr.Flush()
	defer color.Unset()
}

func printFollow(s twitch.FollowChannel) {
	fmt.Fprintln(wr, "ok")
	defer wr.Flush()
}

func printNotification(body string) {
	exec.Command("notify-send", "GoTwitch", body).Start()
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

func urlEncode(str string) string {
	u, _ := url.Parse(str)
	fmt.Fprintln(wr, u.String())
	return u.String()

}

func setSettings(twitchUser, twitchOauthToken, playerName string) {
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
			Padding: 20,
		},
	}
	settingsJson, _ := json.MarshalIndent(settings, "", "\t")
	fmt.Printf(
		"Your settings are located at:\n%s\n\nand look like so:\n%s\n",
		settingsPath,
		settingsJson,
	)
	os.MkdirAll(filepath.Dir(settingsPath), 0775)
	ioutil.WriteFile(settingsPath, settingsJson, 776)
}
