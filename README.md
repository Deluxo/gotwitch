# gotwitch
*Twitch client written in Golang*

It is a fairly simple approach to interacting with Twitch.tv through a command-line interface.
It is written in [Go](https://golang.org/) and is a [free software](https://www.fsf.org/about/what-is-free-software).

## Features

* Stream lookup
    - By game title
    - By type (live, playlist)
    - Subscribed only
* Follow a twitch.tv user
* Open the stream with a player like mpv or vlc
* Game lookup

More features are comig as long as there is demand for it!

Feel free to comment, fork it...or do whatever you please with it!

## Usage
For the first timers, there's unfortunately a setup to go through:

run:
* `gotwitch setup -a` to generate a OAuth token.
    - *It's an encoded login that you will have in plain text in your config file.
    Don't use it if you feel insecure with that!*
* `gotwitch setup -u YOUR_TWITCH.TV_USERNAME -o YOUR_OAUTH_TOKEN -p YOUR_PLAYER_OF_CHOICE`
    - *The recommended player is [mpv](https://mpv.io/), but feel free to experiment with other players as well.*
* `gotwitch watch STREAMER_USERNAME` or `gotwitch streams -b` to get the list of subscribed users.

optional:
* add this to your .zshrc or .bashrc to have tab-completion:
    - If you use zsh: `eval "$(gotwitch --completion-script-zsh)"`
    - If you use bash: `eval "$(gotwitch --completion-script-bash)"`
