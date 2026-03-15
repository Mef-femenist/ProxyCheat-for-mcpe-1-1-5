package utils

import (
	"bytes"
	SkinConverter "github.com/Suremeo/skinconverter"
	"image/png"
	
	"os"
)

func Png2skin(path string) ([]byte, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}

	pngimage, err := png.Decode(bytes.NewBuffer(data))
	if err != nil {
		return nil, false
	}

	return SkinConverter.ImageToSkinData(pngimage), true
}

func Png2skinBytes(b []byte) ([]byte, bool) {
	pngimage, err := png.Decode(bytes.NewBuffer(b))
	if err != nil {
		return nil, false
	}

	return SkinConverter.ImageToSkinData(pngimage), true
}

func Skin2pngBytes(b []byte) ([]byte, bool) {
	img := SkinConverter.SkinDataToImage(b)
	save := bytes.NewBuffer(nil)
	err := png.Encode(save, img)
	if err != nil {
		return nil, false
	}
	return save.Bytes(), true
}

func Skin2png(data string, path string) bool {
	img := SkinConverter.SkinDataToImage([]byte(data))

	save := bytes.NewBuffer(nil)

	err := png.Encode(save, img)
	if err != nil {
		return false
	}

	err = os.WriteFile(path, save.Bytes(), os.ModePerm)
	if err != nil {
		return false
	}
	return true
}
