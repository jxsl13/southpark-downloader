package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/jxsl13/southpark-downloader/utils"
	giturls "github.com/whilp/git-urls"
)

type Config struct {
	YouTubeDLDir string `koanf:"youtube.dl.dir" short:"y" description:"Path to yt-dlp directory"`
	OutDir       string `koanf:"out.dir" short:"o" description:"Output directory"`
	ConfigDir    string `koanf:"config.dir" short:"c" description:"Cache directory"`

	Reinitialize bool `koanf:"reinitialize" short:"i" description:"Re-initialize yt-dlp"`
	DryRun       bool `koanf:"dry.run" short:"d" description:"Dry run: don't download, just print out URLs"`

	RepoUrl string `koanf:"repo.url" short:"r" description:"URL to yt-dlp repository"`
	Branch  string `koanf:"branch" short:"b" description:"Branch to use for yt-dlp"`

	All     bool `koanf:"all" short:"a" description:"Download all episodes"`
	Season  int  `koanf:"season" short:"s" description:"Download all episodes of a season"`
	Episode int  `koanf:"episode" short:"e" description:"Download a specific episode"`

	UserAgent string `koanf:"user.agent" description:"User agent to use for requests"`

	MinRate string `koanf:"min.rate" description:"Minimum download rate"`
}

var rateRegex = regexp.MustCompile(`^\d+[KMG]$`)

func (c *Config) Validate() error {
	if c.All && (c.Season != 0 || c.Episode != 0) {
		return fmt.Errorf("cannot use --all and --season or --episode at the same time")
	}

	if !c.All && c.Season == 0 && c.Episode == 0 {
		return fmt.Errorf("must specify either --all or --season or --episode or --season and --episode")
	}

	if c.Season < 1 {
		return fmt.Errorf("season must be greater than 0")
	}

	if c.Episode < 0 {
		return fmt.Errorf("episode must be greater than or equal to 0")
	}

	foundYtDlDir, err := utils.ExistsDir(c.YouTubeDLDir)
	if err != nil {
		return err
	}

	foundOutDir, err := utils.ExistsDir(c.OutDir)
	if err != nil {
		return err
	}

	foundConfigDir, err := utils.ExistsDir(c.ConfigDir)
	if err != nil {
		return err
	}

	if foundYtDlDir && c.Reinitialize {
		err := os.RemoveAll(c.YouTubeDLDir)
		if err != nil {
			return err
		}
		foundYtDlDir = false
	}

	_, err = giturls.Parse(c.RepoUrl)
	if err != nil {
		return fmt.Errorf("invalid git url: %w", err)
	}

	if !foundYtDlDir {
		err := utils.GitCloneBranch(context.Background(), c.YouTubeDLDir, c.RepoUrl, c.Branch)
		if err != nil {
			return err
		}
	}

	if !foundOutDir {
		err := os.MkdirAll(c.OutDir, 0755)
		if err != nil {
			return err
		}
	}

	if !foundConfigDir {
		err := os.MkdirAll(c.ConfigDir, 0700)
		if err != nil {
			return err
		}
	}

	if !rateRegex.MatchString(c.MinRate) {
		return fmt.Errorf("invalid min rate: %q, must match %s", c.MinRate, rateRegex.String())
	}

	return nil
}

func (c *Config) DBPath() string {
	return filepath.Join(c.ConfigDir, "southpark.db")
}
