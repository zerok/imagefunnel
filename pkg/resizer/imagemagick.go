package resizer

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/rs/zerolog"
)

// ImageMagickResizer implements Resizer using the `magick` CLI.
type ImageMagickResizer struct{}

func (r *ImageMagickResizer) Resize(ctx context.Context, out string, in string, portraitSize string, landscapeSize string) error {
	logger := zerolog.Ctx(ctx)
	bin, err := exec.LookPath("magick")
	if err != nil {
		return err
	}
	logger.Debug().Msgf("%s -> %s", in, out)
	// First, we need to learn the dimensions of the input image
	width, height, err := r.dimensions(ctx, in)
	if err != nil {
		return err
	}
	var cmd *exec.Cmd
	if width > height {
		cmd = exec.CommandContext(ctx, bin, in, "-resize", landscapeSize, out)
	} else {
		cmd = exec.CommandContext(ctx, bin, in, "-resize", portraitSize, out)
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
func (r *ImageMagickResizer) dimensions(ctx context.Context, path string) (int64, int64, error) {
	var output bytes.Buffer
	cmd := exec.Command("magick", "identify", path)
	cmd.Stdout = &output
	if err := cmd.Run(); err != nil {
		return 0, 0, err
	}
	elems := strings.Split(output.String(), " ")
	if len(elems) < 3 {
		return 0, 0, fmt.Errorf("Failed to retrieve dimensions of %s", path)
	}
	elems = strings.Split(elems[2], "x")
	if len(elems) < 2 {
		return 0, 0, fmt.Errorf("Failed to retrieve dimensions of %s", path)
	}
	width, err := strconv.ParseInt(elems[0], 10, 64)
	if err != nil {
		return 0, 0, err
	}
	height, err := strconv.ParseInt(elems[1], 10, 64)
	if err != nil {
		return 0, 0, err
	}
	return width, height, nil
}

func NewImageMagick() Resizer {
	return &ImageMagickResizer{}
}
