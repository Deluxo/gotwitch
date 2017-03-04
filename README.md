# gotwitch
*Twitch client written in Golang*

It is a fairly simple approach to interacting with Twitch.tv through a command-line interface.
It is written in [Go](https://golang.org/) and is a [free software](https://www.fsf.org/about/what-is-free-software).

## Help

```
usage: gotwitch [<flags>] <command> [<args> ...]

A command-line twitch.tv application written in Golang.

Flags:
      --help          Show context-sensitive help (also try --help-long and
                      --help-man).
  -d, --padding=25    output column padding width
  -r, --url           print channels full URL for lazy people
  -w, --watch         watch the stream through a given player
  -p, --player="mpv"  player to use for watching a stream
  -f, --follow        follow the streamer
  -n, --notify        notify if the streamer comes online
  -u, --unfollow      unfollow the streamer
  -q, --search        search for the streamer with a given name
  -l, --ls            list the streamers
  -b, --subscribed    filter out only subscribed streamers
  -g, --game          print the game a streamer is playing
  -s, --status        print the streamer's status

Args:
  [<channel>]  channel of a streamer

Commands:
  help [<command>...]
    Show help.


  streamer [<flags>] [<channel>]
    Do actions related to a streamer

    -r, --url           print channels full URL for lazy people
    -w, --watch         watch the stream through a given player
    -p, --player="mpv"  player to use for watching a stream
    -f, --follow        follow the streamer
    -n, --notify        notify if the streamer comes online
    -u, --unfollow      unfollow the streamer
    -q, --search        search for the streamer with a given name
    -l, --ls            list the streamers
    -b, --subscribed    filter out only subscribed streamers
    -g, --game          print the game a streamer is playing
    -s, --status        print the streamer's status

  game [<flags>] [<title>]
    Do actions related to a game

    -o, --offset=0  game list view starting point
    -l, --limit=10  game list view length

  setup [<flags>]
    setup procedure

    --username=USERNAME        twitch.tv channel
    --access-token="generate"  a generated access token provided by twitch.tv
    --player=PLAYER            video player to use for stream watching by
                               default


```
