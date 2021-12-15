package main

import (
	"bmp-stega/bmpio"
	"encoding/binary"
	"flag"
	"fmt"
	"image"
	"image/color"
	"os"

	"github.com/sirupsen/logrus"
)

func main() {
	in := flag.String("input", "", "input image")
	out := flag.String("output", "", "output image")
	decode := flag.Bool("d", false, "decode?")
	dbg := flag.Bool("dbg", false, "debug")

	flag.Parse()

	if *dbg {
		logrus.Info("Running in debug mode")
		logrus.SetLevel(logrus.DebugLevel)
	}

	message := []uint8{0, 1, 1, 0, 1, 1, 1, 1, 1, 0, 0, 1, 1}

	logrus.Debug("Debugging with default message:\n")
	logrus.Debug(message)

	if *in == "" {
		logrus.Fatal(fmt.Errorf("input file incorrect"))
	}

	if !*decode { // action = false == encode
		if *out == "" {
			logrus.Fatal(fmt.Errorf("output file incorrect"))
		}

		img, err := bmpio.ReadImage(*in)
		if err != nil {
			logrus.Fatal(err)
		}

		poisonedImage, err := injectMessage(img, message)
		if err != nil {
			logrus.Fatal(err)
		}

		if err := bmpio.WriteImage(*out, poisonedImage); err != nil {
			logrus.Fatal(err)
		}

		if file, err := os.ReadFile(*out); err != nil {
			logrus.Fatal(err)
		} else {
			file = injectMessageSize(file, uint32(len(message)))

			if err := os.WriteFile(*out, file, 0664); err != nil {
				logrus.Fatal(err)
			}
		}
	} else {
		file, err := os.ReadFile(*in)
		if err != nil {
			logrus.Fatal(err)
		}

		var newMsg []uint8
		if detectInjection(file) {
			img, err := bmpio.ReadImage(*in)
			if err != nil {
				logrus.Fatal(err)
			}

			newMsg = extractMessage(img, extractMessageSize(file))
		}

		logrus.Println("Extracted from file:")
		logrus.Println(newMsg)
	}
}

func stringToBin(s string) (binString string) {
	for _, c := range s {
		binString = fmt.Sprintf("%s%b", binString, c)
	}
	return
}

func extractMessage(img image.Image, length uint32) []uint8 {
	var result []uint8

	bounds := img.Bounds()

	for k, i := 0, 0; i < bounds.Dx(); i++ {
		for j := 0; j < bounds.Dy(); j++ {
			clr := img.At(i, j)
			// extraction
			if uint32(k) < length {
				r, _, _, _ := clr.RGBA()
				result = append(result, getBit(r))
				k++
			} else {
				break
			}
		}
	}

	return result
}

func injectMessage(img image.Image, message []uint8) (image.Image, error) {
	bounds := img.Bounds()

	newImg := image.NewRGBA(image.Rectangle{Min: image.Point{}, Max: image.Point{bounds.Dx(), bounds.Dy()}})

	for k, i := 0, 0; i < bounds.Dx(); i++ {
		for j := 0; j < bounds.Dy(); j++ {
			clr := img.At(i, j)
			// injection
			if k < len(message) {
				oldR, oldG, oldB, oldA := clr.RGBA()

				test := color.RGBA{
					R: uint8(setBit(oldR, message[k])),
					G: uint8(oldG),
					B: uint8(oldB),
					A: uint8(oldA),
				}

				newImg.Set(i, j, test)
				k++
			} else {
				newImg.Set(i, j, clr)
			}
		}
	}

	return newImg, nil
}

func setBit(integer uint32, bit uint8) uint32 { // todo tests
	if bit >= 1 {
		return integer | (1 << 0)
	}

	mask := ^(1 << 0)

	return integer & uint32(mask)
}

func getBit(integer uint32) uint8 {
	return uint8(integer % 2)
}

func extractMessageSize(file []byte) uint32 {
	if detectInjection(file) {
		bs := file[len(file)-5 : len(file)-1]
		return binary.LittleEndian.Uint32(bs)
	}

	return 0
}

func injectMessageSize(file []byte, size uint32) []byte {
	bs := make([]byte, 4)
	binary.LittleEndian.PutUint32(bs, size)

	file = append(file, 0x55)
	file = append(file, bs...)
	file = append(file, 0x55)

	return file
}

func detectInjection(file []byte) bool {
	if file[len(file)-1] == 0x55 && file[len(file)-6] == 0x55 {
		return true
	}

	return false
}
