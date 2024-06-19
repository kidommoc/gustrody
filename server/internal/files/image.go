package files

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"time"

	"github.com/kidommoc/gustrody/internal/logging"
	"github.com/kidommoc/gustrody/internal/utils"
)

const (
	TYPE_JPEG FileType = "jpeg"
	TYPE_PNG  FileType = "png"
)

func (service *FileService) StoreImage(user string, buf []byte) (path string, err utils.Error) {
	logger := logging.Get()
	img := File{Uploader: user}

	// digest []byte for filename
	// conflict may not resolved
	h := sha1.New()
	// add datetime string to avoid conflict
	h.Write([]byte(time.Now().Format(time.RFC3339)))
	h.Write(buf)
	digest := hex.EncodeToString(h.Sum(nil))
	img.Filename = digest

	// check image type
	buffer := bytes.NewBuffer(buf)
	_, t, e := image.Decode(buffer)
	if e != nil {
		err = newErr(ErrFile, "decode: "+e.Error())
		logger.Error("[Files.Image] Cannot decode file to image", err)
		return "", err
	}
	switch t {
	case "jpg":
	case "jpeg":
		img.Ext = TYPE_JPEG
	case "png":
		img.Ext = TYPE_PNG
	default:
		err = newErr(ErrFile, "type: "+t)
		logger.Error("[Files.Image] Wrong file type", err)
		return "", err
	}

	dst := fmt.Sprintf("%s/%s.%s",
		service.imgDir,
		img.Filename,
		string(img.Ext),
	)

	if e := service.storeFile(dst, buf); e != nil {
		// handle error
		logger.Error("[Files.Image] Cannot store image", e)
		return "", e
	}

	return fmt.Sprintf("/imgs/%s.%s", img.Filename, img.Ext), nil
}
