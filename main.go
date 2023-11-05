package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
	"github.com/jxsl13/southpark-downloader/config"
	"github.com/jxsl13/southpark-downloader/utils"
	"github.com/spf13/cobra"
	_ "modernc.org/sqlite"
)

func main() {
	err := NewRootCmd().Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func NewRootCmd() *cobra.Command {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)

	rootContext := rootContext{Ctx: ctx}

	// cmd represents the run command
	cmd := &cobra.Command{
		Use:   "southpark-downloader",
		Short: "download new southpark episodes",
		RunE:  rootContext.RunE,
		Args:  cobra.ExactArgs(0),
		PostRunE: func(cmd *cobra.Command, args []string) error {

			cancel()
			return nil
		},
	}

	// register flags but defer parsing and validation of the final values
	cmd.PreRunE = rootContext.PreRunE(cmd)

	// register flags but defer parsing and validation of the final values
	cmd.AddCommand(NewCompletionCmd(cmd.Name()))
	return cmd
}

type rootContext struct {
	Ctx    context.Context
	Config *config.Config
	DB     *sql.DB
}

func (c *rootContext) PreRunE(cmd *cobra.Command) func(cmd *cobra.Command, args []string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "./"
	}

	c.Config = &config.Config{
		Reinitialize: false,
		YouTubeDLDir: "./yt-dlp",
		OutDir:       "./downloads",
		ConfigDir:    filepath.Join(home, ".config", "southpark-downloader"),
		RepoUrl:      "https://github.com/yt-dlp/yt-dlp.git",
		Branch:       "2023.03.04",
		MinRate:      "1M",

		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.0.0 Safari/537.36",
	}

	runParser := config.RegisterFlags(c.Config, true, cmd)

	return func(cmd *cobra.Command, args []string) error {

		err := runParser()
		if err != nil {
			return err
		}

		if !utils.IsApplicationAvailable(c.Ctx, "ffmpeg") {
			return fmt.Errorf("%w: ffmpeg", utils.ErrApplicationNotFound)
		}

		err = c.InitDB()
		if err != nil {
			return err
		}

		return nil
	}
}

func (c *rootContext) PostRunE(cmd *cobra.Command, args []string) error {
	return c.CloseDB()
}

var (
	// /folgen/940f8z/south-park-cartman-und-die-analsonde-staffel-1-ep-1
	// /episodes/940f8z/south-park-cartman-gets-an-anal-probe-season-1-ep-1
	episodeUrlRegex = regexp.MustCompile(`/[a-z]+/[0-9a-z]+/south-park-[0-9a-z-]+-[a-z]+-[0-9]+-[a-z]+-[0-9]+$`)
)

func (c *rootContext) RunE(cmd *cobra.Command, args []string) (err error) {
	err = c.CollectUrls()
	if err != nil {
		return fmt.Errorf("failed to collect urls: %w", err)
	}

	return c.Download(c.Config.Season, c.Config.Episode)
}

func (c *rootContext) CollectUrls() error {
	var (
		ctx       = c.Ctx
		userAgent = c.Config.UserAgent
	)

	startUrl, err := c.Last()
	if err != nil {
		if !errors.Is(err, ErrNotFound) {
			return err
		}
		startUrl, err = StartingUrl(ctx)
		if err != nil {
			return err
		}
	}

	co := NewCollector(ctx, userAgent)

	co.OnScraped(func(r *colly.Response) {
		if len(r.Body) == 0 {
			r.Request.Visit(r.Request.URL.String())
		}
	})

	// prevent skipping first request
	skippable := false
	co.OnRequest(func(r *colly.Request) {
		visited, _ := c.Visited(r.URL.String())
		if !skippable {
			skippable = true
		} else if visited {
			fmt.Println("Skipping:", r.URL.String())
			r.Abort()
			return
		}
		fmt.Println("Getting:", r.URL.String())
	})

	co.OnResponse(func(r *colly.Response) {
		fmt.Println("Got:", r.Request.URL.String())

	})

	co.OnHTML("html", func(e *colly.HTMLElement) {
		url := e.Request.URL.String()
		meta := e.DOM.Find("meta[property]")

		var (
			title         string
			seasonNumber  int
			episodeNumber int
			description   string
			imageUrl      string
			contentDate   time.Time
		)

		cnt := 0
		meta.Each(func(i int, s *goquery.Selection) {
			val, _ := s.Attr("property")
			switch val {
			case "search:episodeTitle":
				title, _ = s.Attr("content")
				cnt++
			case "search:seasonNumber":
				seasonNumber, _ = strconv.Atoi(s.AttrOr("content", "0"))
				cnt++
			case "search:episodeNumber":
				episodeNumber, _ = strconv.Atoi(s.AttrOr("content", "0"))
				cnt++
			case "og:description":
				description, _ = s.Attr("content")
				cnt++
			case "og:image":
				imageUrl, _ = s.Attr("content")
				cnt++
			case "og:video:release_date":
				const contentLayout = "2006-01-02T15:04:05.000Z"
				contentDate, _ = time.Parse(contentLayout, s.AttrOr("content", "1970-01-01T00:00:00.000Z"))
				cnt++
			}
		})

		if cnt < 6 {
			fmt.Fprintf(os.Stderr, "failed to parse meta tags: %v\n", meta)
			return
		}

		err = c.Insert(
			title,
			seasonNumber,
			episodeNumber,
			url,
			description,
			imageUrl,
			contentDate,
		)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to insert episode: %v\n", err)
			e.Request.Abort()
			return
		}

	})

	co.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")

		if e.Request.URL.Path == link {
			return
		}

		if e.Request.URL.String() == link {
			return
		}

		// broken links
		_, err := url.Parse(link)
		if err != nil {
			return
		}

		if episodeUrlRegex.MatchString(link) {
			visited, err := c.Visited(link)
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to check if episode was visited: %v\n", err)
				e.Request.Abort()
				return
			}

			if !visited {
				e.Request.Visit(link)
			}
		}
	})

	err = co.Visit(startUrl)
	if err != nil {
		return err
	}

	return nil
}

func (c *rootContext) Download(season, episode int) error {
	videos, err := c.Videos(season, episode)
	if err != nil {
		return err
	}

	for _, v := range videos {
		err = c.DownloadVideo(v)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *rootContext) Videos(season, episode int) ([]Video, error) {
	if season == 0 && episode == 0 {
		return c.All()
	}

	if episode == 0 {
		return c.Season(season)
	}

	v, err := c.Episode(season, episode)
	if err != nil {
		return nil, err
	}
	return []Video{v}, nil
}

func (c *rootContext) DownloadVideo(v Video) (err error) {

	outDir := filepath.Join(c.Config.OutDir, v.SeasonString())
	err = os.MkdirAll(outDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	exe := "yt-dlp"
	if runtime.GOOS == "windows" {
		exe += ".cmd"
	} else {
		exe += ".sh"
	}

	cmd := filepath.Join(c.Config.YouTubeDLDir, exe)
	absCmd, err := filepath.Abs(cmd)
	if err != nil {
		return err
	}

	if c.Config.DryRun {
		fmt.Println("Would download:", v.Url)
		return nil
	}

	err = utils.ExecutePathApplication(
		c.Ctx,
		outDir,
		absCmd,
		"--concurrent-fragments",
		strconv.Itoa(runtime.NumCPU()),
		"--throttled-rate",
		c.Config.MinRate,
		"--output",
		v.Format(),
		v.Url,
	)
	if err != nil {
		return err
	}
	return nil
}
