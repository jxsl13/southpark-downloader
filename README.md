# South Park Downloader


This is a download utility that is highly inspired by https://github.com/xypwn/southpark-de-downloader
which does not work that well on macOS.

- creates a sqlite3 database as index

```shell
go install github.com/jxsl13/southpark-downloader@latest
```

## Requirements

- git (for downloading yt-dlp)
- python3 (uses yt-dlp)
- ffmpeg (transcoding/stream decryption)


## Usage

```text
$ southpark-downloader --help
Environment variables:
  SPDL_YOUTUBE_DL_DIR    Path to yt-dlp directory (default: "./yt-dlp")
  SPDL_OUT_DIR           Output directory (default: "./downloads")
  SPDL_CONFIG_DIR        Cache directory (default: "/Users/john/.config/southpark-downloader")
  SPDL_REINITIALIZE      Re-initialize yt-dlp (default: "false")
  SPDL_DRY_RUN           Dry run: don't download, just print out URLs (default: "false")
  SPDL_REPO_URL          URL to yt-dlp repository (default: "https://github.com/yt-dlp/yt-dlp.git")
  SPDL_BRANCH            Branch to use for yt-dlp (default: "2023.03.04")
  SPDL_ALL               Download all episodes (default: "false")
  SPDL_SEASON            Download all episodes of a season (default: "0")
  SPDL_EPISODE           Download a specific episode (default: "0")
  SPDL_USER_AGENT        User agent to use for requests (default: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.0.0 Safari/537.36")
  SPDL_MIN_RATE          Minimum download rate (default: "1M")

Usage:
  southpark-downloader [flags]
  southpark-downloader [command]

Available Commands:
  completion  Generate completion script
  help        Help about any command

Flags:
  -a, --all                     Download all episodes
  -b, --branch string           Branch to use for yt-dlp (default "2023.03.04")
  -c, --config-dir string       Cache directory (default "~/.config/southpark-downloader")
  -d, --dry-run                 Dry run: don't download, just print out URLs
  -e, --episode int             Download a specific episode
  -h, --help                    help for southpark-downloader
      --min-rate string         Minimum download rate (default "1M")
  -o, --out-dir string          Output directory (default "./downloads")
  -i, --reinitialize            Re-initialize yt-dlp
  -r, --repo-url string         URL to yt-dlp repository (default "https://github.com/yt-dlp/yt-dlp.git")
  -s, --season int              Download all episodes of a season
      --user-agent string       User agent to use for requests (default "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.0.0 Safari/537.36")
  -y, --youtube-dl-dir string   Path to yt-dlp directory (default "./yt-dlp")

Use "southpark-downloader [command] --help" for more information about a command.
```


Example:

```
# download all seasons
southpark-downloader -a

# download season 26
southpark-downloader -s 26

# download episode 1 of season 26
southpark-downloader -s 26 -e 1
```
