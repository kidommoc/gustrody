package files

import (
	"fmt"
	"io/fs"
	"os"

	"github.com/kidommoc/gustrody/internal/config"
	"github.com/kidommoc/gustrody/internal/logging"
	"github.com/kidommoc/gustrody/internal/utils"
)

type FileType string

type File struct {
	Filename string   `json:"filename"`
	Ext      FileType `json:"ext"`
	Uploader string   `json:"uploader"`
}

type FileService struct {
	imgDir string
}

func NewFileService(c ...config.Config) *FileService {
	var cfg config.Config
	if len(c) == 0 {
		cfg = config.Get()
	} else {
		cfg = c[0]
	}
	return &FileService{
		imgDir: cfg.ImgDir,
	}
}

func (service *FileService) storeFile(dst string, data []byte) utils.Error {
	logger := logging.Get()
	f, e := os.OpenFile(dst, os.O_CREATE, fs.ModePerm)
	if e != nil {
		err := newErr(ErrFsInternal, e.Error())
		msg := fmt.Sprintf("[Files] Cannot open file to store: %s", dst)
		logger.Error(msg, err)
		return err
	}
	defer f.Close()

	n, e := f.Write(data)
	if e != nil {
		err := newErr(ErrFsInternal, e.Error())
		msg := fmt.Sprintf("[Files] Cannot write to file: %s", dst)
		logger.Error(msg, err)
		return err
	}
	logger.Info("[File] Store a file.",
		"dst", dst,
		"size", n,
	)
	return nil
}
