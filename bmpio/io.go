package bmpio

import (
	"image"
	"os"

	"golang.org/x/image/bmp"
)

func ReadImage(filename string) (image.Image, error) {
	imageFile, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	return bmp.Decode(imageFile)
}

func WriteImage(filename string, image image.Image) error {
	if file, err := os.Create(filename); err != nil {
		return err
	} else {
		if imgErr := bmp.Encode(file, image); imgErr != nil {
			return imgErr
		}
	}

	return nil
}
