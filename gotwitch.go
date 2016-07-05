package main

import (
	"encoding/json"
	"fmt"
	"github.com/deluxo/gotwitchlib"
	"net/url"
	"github.com/fatih/color"
	"gopkg.in/alecthomas/kingpin.v2"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"strconv"
)

type User struct {
	Username   string
	OauthToken string
}
type Player struct {
	Quality string
	Name    string
}
type Options struct {
	Game   bool
	Status bool
}
type Settings struct {
	User    User
	Player  Player
	Options Options
}

var (
	usr, _       = user.Current()
	settingsPath = usr.HomeDir + "/.config/gotwitch/config.json"

	app = kingpin.New(
		"gotwitch",
		"A command-line twitch.tv application written in Golang.",
	)

	debug = app.Flag(
		"debug",
		"Enable debug mode.",
	).Short('d').Bool()

	// online-streams
	onlineSubs = app.Command(
		"online-subs",
		"Look up online subscriptions currently streaming.",
	)

	// streams
	streams = app.Command(
		"streams",
		"Look up online streams.",
	)

	streamsGame = streams.Arg(
		"game",
		"Game title, like dota2",
	).String()

	streamsType = streams.Arg(
		"type",
		"Stream type, e.g. live, all, or playlist",
	).String()

	streamsLimit = streams.Arg(
		"limit",
		"Number of streams to print.",
	).Int()

	streamsOffset = streams.Arg(
		"offset",
		"Pagination offset for given limit.",
	).Int()

	// watch
	watch = app.Command(
		"watch",
		"Watch the stream.",
	)

	streamer = watch.Arg(
		"streamer",
		"Streamer name.",
	).Required().String()

	// top-games
	topGames = app.Command(
		"top-games",
		"Get the list of currently most played games.",
	)

	limit = topGames.Arg(
		"limit",
		"Wanted list size of a request.",
	).Int()

	offset = topGames.Arg(
		"offset",
		"Pagination offset of a page (default page size is 10).",
	).Int()

	// setup
	setup = app.Command(
		"setup",
		"create a config file with required values",
	)

	twitchUser = setup.Arg(
		"username",
		"Twitch.tv username.",
	).Required().String()

	twitchOauthToken = setup.Arg(
		"oauth",
		"Twitch.tv oAuthToken (must be generated first).",
	).Required().String()

	playerName = setup.Arg(
		"player",
		"Player command to be used along with livestreamer, like mpv or vlc.",
	).Required().String()

	playerQuality = setup.Arg(
		"quality",
		"Default stream quality to be used, like best, worst, 720p, 480p.",
	).Required().String()

	// flags
	notify = app.Flag(
		"notify",
		"Print the output through the notification instead of std out.",
	).Short('n').Bool()

	statusFlag = app.Flag(
		"status",
		"Include stream status into output.",
	).Short('s').Bool()

	gameFlag = app.Flag(
		"game",
		"Include streamded game title into output.",
	).Short('g').Bool()
)

func main() {
	kingpin.CommandLine.HelpFlag.Short('h')
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {

	case onlineSubs.FullCommand():
		s := getSettings()
		onlineSubs := twitch.GetOnlineSubs(s.User.OauthToken)
		if *notify == true {
			var notification string
			for k, v := range onlineSubs.Streams {
				notification += strconv.Itoa(k+1) + ". " + v.Channel.Name + "\n"
			}
			exec.Command("notify-send", "GoTwitch", notification).Start()
		} else {
			for _, v := range onlineSubs.Streams {
				printStream(v.Channel, statusFlag, gameFlag)
			}
		}

	case streams.FullCommand():
		streams := twitch.GetStreams(url.QueryEscape(*streamsGame), *streamsType, *streamsLimit, *streamsOffset)
		if *notify == true {
			var notification string
			for k, v := range streams.Streams {
				notification += strconv.Itoa(k+1) + ". " + v.Channel.Name + "\n"
			}
			exec.Command("notify-send", "GoTwitch", notification).Start()
		} else {
			for _, v := range streams.Streams {
				printStream(v.Channel, statusFlag, gameFlag)
			}
		}

	case topGames.FullCommand():
		topGames := twitch.GetTopGames(limit, offset)
		if *notify == true {
			var notification string
			for k, v := range topGames.Top {
				notification += strconv.Itoa(k+1) + ". " + v.Game.Name + "\n"
			}
			exec.Command("notify-send", "GoTwitch", notification).Start()
		} else {
			for _, v := range topGames.Top {
				printGame(v.Game)
			}
		}

	case watch.FullCommand():
		s := getSettings()
		exec.Command(
			"livestreamer",
			"-Q",
			twitch.TwitchUrl+*streamer,
			s.Player.Quality,
			"--player",
			s.Player.Name).Start()
	case setup.FullCommand():
		setSettings(*twitchUser, *twitchOauthToken, *playerName, *playerQuality)
	}
}

func getSettings() Settings {
	var settings Settings
	settingsFile, err := os.Open(settingsPath)
	if err != nil {
		fmt.Println("No config file was found.\nConsider running setup command first.")
	}
	fileStream, _ := ioutil.ReadAll(settingsFile)
	json.Unmarshal(fileStream, &settings)
	return settings
}

func printGame(s twitch.Game) {
	game := color.New(color.Bold, color.FgHiRed).SprintFunc()
	lineColored := game(s.Name) + "\n"
	println(lineColored)

	defer color.Unset()
}

func printStream(s twitch.Channel, showFlag *bool, gameFlag *bool) {
	nick := color.New(color.FgHiBlue).SprintFunc()
	status := color.New(color.FgHiWhite).SprintFunc()
	game := color.New(color.Bold, color.FgHiRed).SprintFunc()
	lineColored := nick(s.Name) + "\n"
	if *showFlag == true {
		lineColored += status(s.Status) + "\n"
	}
	if *gameFlag == true {
		lineColored += game(s.Game) + "\n"
	}
	println(lineColored)

	defer color.Unset()
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

func setSettings(twitchUser, twitchOauthTokeno, playerName, playerQuality string) {
	settings := Settings{
		User: User{
			Username:   twitchUser,
			OauthToken: *twitchOauthToken,
		},
		Player: Player{
			Quality: playerQuality,
			Name:    playerName,
		},
		Options: Options{
			Game:   false,
			Status: false,
		},
	}
	settingsJson, _ := json.MarshalIndent(settings, "", "\t")
	fmt.Printf(
		"Your settings are located at:\n%s\n\nand look like so:\n%s\n",
		settingsPath,
		settingsJson,
	)
	ioutil.WriteFile(settingsPath, settingsJson, 776)
}
