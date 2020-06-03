package resizer

import "context"

// Resizer a struct that allows the resizing of a given input image into an
// output image with specified dimensions.
type Resizer interface {
	Resize(ctx context.Context, outputpath string, inputpath string, portraitSize string, landscapeSize string) error
}
