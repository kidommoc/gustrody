package files

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
)

const (
	TYPE_JPEG FileType = "jpeg"
	TYPE_PNG  FileType = "png"
)

func (service *FileService) StoreImage(user string, buf []byte) (url string, mediaType string, err error) {
	logger := service.lg
	img := File{Uploader: user}

	// digest []byte for filename
	// conflict may not resolved
	img.Filename = digestToHex(buf)

	// check image type
	buffer := bytes.NewBuffer(buf)
	_, t, e := image.Decode(buffer)
	if e != nil {
		logger.Error("[Files.Image] Cannot decode file to image", e)
		return "", "", ErrFile
	}
	switch t {
	case "jpg":
	case "jpeg":
		img.Ext = TYPE_JPEG
	case "png":
		img.Ext = TYPE_PNG
	default:
		logger.Error("[Files.Image] Wrong file type", errors.New("expect jpeg/png, got "+t))
		return "", "", ErrFile
	}

	dst := fmt.Sprintf("%s/%s.%s",
		service.imgDir,
		img.Filename,
		string(img.Ext),
	)

	if e := service.storeFile(dst, buf); e != nil {
		logger.Error("[Files.Image] Cannot store image", e)
		return "", "", ErrFsInternal
	}

	url = fmt.Sprintf("%s/imgs/%s.%s", service.site, img.Filename, img.Ext)
	mediaType = "image/" + string(img.Ext)
	return url, mediaType, nil
}
