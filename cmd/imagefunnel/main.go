package main

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/rs/zerolog"
	"github.com/spf13/pflag"
	"github.com/zerok/imagefunnel/pkg/profile"
	"github.com/zerok/imagefunnel/pkg/resizer"
	"gopkg.in/yaml.v2"
)

func loadProfile(fpath string) (*profile.Profile, error) {
	var prof profile.Profile
	fp, err := os.Open(fpath)
	if err != nil {
		return nil, err
	}
	defer fp.Close()
	err = yaml.NewDecoder(fp).Decode(&prof)
	return &prof, err
}
func main() {
	var logger zerolog.Logger
	var profilePaths []string
	pflag.StringSliceVar(&profilePaths, "profile", []string{}, "Profile documents that should be processed for the specified input")
	pflag.Parse()

	logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger()

	ctx := logger.WithContext(context.Background())

	profiles := make([]*profile.Profile, 0, len(profilePaths))

	for _, p := range profilePaths {
		prof, err := loadProfile(p)
		if err != nil {
			logger.Fatal().Err(err).Msgf("Failed to parse profile from %s", p)
		}
		profiles = append(profiles, prof)
	}
	for _, arg := range pflag.Args() {
		files, err := ioutil.ReadDir(arg)
		if err != nil {
			logger.Fatal().Err(err).Msgf("Failed to read directory %s", arg)
		}
		if err := processFiles(ctx, arg, files, profiles); err != nil {
			logger.Fatal().Err(err).Msgf("Failed to process directory %s", arg)
		}
	}
}

func processFiles(ctx context.Context, basedir string, files []os.FileInfo, profiles []*profile.Profile) error {
	for _, f := range files {
		ipath := filepath.Join(basedir, f.Name())
		for _, p := range profiles {
			if p.Matches(f.Name()) {
				opath, err := p.CalculateTargetFilename(f.Name())
				if err != nil {
					return err
				}
				opath = filepath.Join(basedir, opath)
				_, err = os.Stat(opath)
				if os.IsNotExist(err) {
					rez := resizer.NewImageMagick()
					if err := rez.Resize(ctx, opath, ipath, p.Target.PortraitSize, p.Target.LandscapeSize); err != nil {
						return err
					}
					continue
				}
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}
