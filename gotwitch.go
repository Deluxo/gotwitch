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
	wr           = bufio.NewWriter(os.Stdout)
	usr, _       = user.Current()
	settingsPath = usr.HomeDir + "/.config/gotwitch/config.json"

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
	twitchUser       = setup.Flag("username", "Twitch.tv username.").Required().String()
	twitchOauthToken = setup.Flag("oauth", "Twitch.tv oAuthToken (must be generated first).").Required().String()
	playerName       = setup.Flag("player", "Player command to be used along with livestreamer, like mpv or vlc.").Required().String()
	playerQuality    = setup.Flag("quality", "Default stream quality to be used, like best, worst, 720p, 480p.").Required().String()
	// flags
)

func main() {
	kingpin.CommandLine.HelpFlag.Short('h')
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {

	case streams.FullCommand():
		s := getSettings()
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
			for _, v := range twitch.GetStreams(*streamsGame, *streamsType, *streamsLimit, *streamsOffset).Streams {
				printStream(v.Channel, streamsPrintStatus, streamsPrintGame)
			}
		}

	case follow.FullCommand():
		s := getSettings()
		response := twitch.Follow(s.User.OauthToken, s.User.Username, *followStreamer, *followNotification)
		printFollow(response)

	case topGames.FullCommand():
		games := twitch.GetTopGames(limit, offset)
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

func listChannels() []string {
	s := getSettings()

	subStreams := twitch.GetLiveSubs(
		s.User.OauthToken).Streams

	popularStreams := twitch.GetStreams(
		*streamsGame,
		*streamsType,
		*streamsLimit,
		*streamsOffset,
	).Streams
	list := make([]string, 0)
	for _, v := range subStreams {
		//fmt.Println(v.Channel.Name)
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
		fmt.Fprintln(wr, "No config file was found.\nConsider running setup command first.")
	}
	fileStream, _ := ioutil.ReadAll(settingsFile)
	json.Unmarshal(fileStream, &settings)
	return settings
}

func printGame(s twitch.Game) {
	game := color.New(color.Bold, color.FgHiRed).SprintFunc()
	lineColored := game(s.Name) + "\n"
	fmt.Fprintln(wr, lineColored)

	defer wr.Flush()
	defer color.Unset()
}

func printStream(s twitch.Channel, showFlag *bool, gameFlag *bool) {
	nick := color.New(color.FgHiBlue).SprintFunc()
	status := color.New(color.FgHiWhite).SprintFunc()
	game := color.New(color.Bold, color.FgHiRed).SprintFunc()
	lineColored := nick(s.Name) + "    "
	if *showFlag == true {
		lineColored += status(s.Status) + "\n"
	}
	if *gameFlag == true {
		lineColored += game(s.Game)
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
