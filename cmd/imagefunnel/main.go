package main

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/minio/minio-go/v6"
	"github.com/rs/zerolog"
	"github.com/spf13/pflag"
	"github.com/zerok/imagefunnel/pkg/profile"
	"github.com/zerok/imagefunnel/pkg/resizer"
	"gopkg.in/yaml.v2"
)

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

	if len(pflag.Args()) > 0 {
		logger.Info().Msg("Processing local files...")

		for _, arg := range pflag.Args() {
			files, err := ioutil.ReadDir(arg)
			if err != nil {
				logger.Fatal().Err(err).Msgf("Failed to read directory %s", arg)
			}
			if err := processFiles(ctx, arg, files, profiles); err != nil {
				logger.Fatal().Err(err).Msgf("Failed to process directory %s", arg)
			}
		}
		return
	}

	endpoint := mustHaveEnvironment(ctx, "IMAGEFUNNEL_S3_ENDPOINT")
	accessKeyID := mustHaveEnvironment(ctx, "IMAGEFUNNEL_S3_ACCESS_KEY_ID")
	secret := mustHaveEnvironment(ctx, "IMAGEFUNNEL_S3_SECRET_ACCESS_KEY")
	bucketName := mustHaveEnvironment(ctx, "IMAGEFUNNEL_S3_BUCKET")

	client, err := minio.New(endpoint, accessKeyID, secret, true)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create client for data store.")
	}
	done := make(chan struct{})
	defer close(done)
	existingFiles := make(map[string]struct{})
	for o := range client.ListObjectsV2(bucketName, "photos/", true, done) {
		existingFiles[o.Key] = struct{}{}
	}

	for f := range existingFiles {
		for _, p := range profiles {
			if p.Matches(f) {
				o, err := p.CalculateTargetFilename(f)
				if err != nil {
					logger.Fatal().Err(err).Msgf("Failed to calculate output for %s", f)
				}
				if _, ok := existingFiles[o]; !ok {
					logger.Info().Msgf("Generating %s -> %s", f, o)
					tmpIn := "tmp-in" + filepath.Ext(f)
					tmpOut := "tmp-out" + filepath.Ext(f)
					if err := client.FGetObjectWithContext(ctx, bucketName, f, tmpIn, minio.GetObjectOptions{}); err != nil {
						logger.Fatal().Err(err).Msgf("Failed to download %s", f)
					}
					r := resizer.NewImageMagick()
					if err := r.Resize(ctx, tmpOut, tmpIn, p.Target.PortraitSize, p.Target.LandscapeSize); err != nil {
						logger.Fatal().Err(err).Msgf("Failed to resize %s", f)
					}
					if _, err := client.FPutObjectWithContext(ctx, bucketName, o, tmpOut, minio.PutObjectOptions{
						UserMetadata: map[string]string{
							"x-amz-acl": "public-read",
						},
					}); err != nil {
						logger.Fatal().Err(err).Msgf("Failed to upload %s", o)
					}
				} else {
					logger.Debug().Msgf("Skipping %s -> %s", f, o)
				}
			}
		}
	}
}

func mustHaveEnvironment(ctx context.Context, name string) string {
	logger := zerolog.Ctx(ctx)
	value, found := os.LookupEnv(name)
	if !found {
		logger.Fatal().Msgf("Environment variable %s not set!", name)
	}
	return value
}

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
