package files

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"time"

	"github.com/kidommoc/gustrody/internal/config"
	"github.com/kidommoc/gustrody/internal/logging"
)

type FileType string

type File struct {
	Filename string   `json:"filename"`
	Ext      FileType `json:"ext"`
	Uploader string   `json:"uploader"`
}

type FileService struct {
	lg     logging.Logger
	site   string
	imgDir string
}

func NewService(cfg config.Config, lg logging.Logger) *FileService {
	return &FileService{
		lg:     lg,
		site:   cfg.Site,
		imgDir: cfg.ImgDir,
	}
}

func (service *FileService) storeFile(dst string, data []byte) error {
	logger := service.lg
	f, e := os.OpenFile(dst, os.O_CREATE, fs.ModePerm)
	if e != nil {
		msg := fmt.Sprintf("[Files] Cannot open file to store: %s", dst)
		logger.Error(msg, e)
		return ErrFsInternal
	}
	defer f.Close()

	n, e := f.Write(data)
	if e != nil {
		msg := fmt.Sprintf("[Files] Cannot write to file: %s", dst)
		logger.Error(msg, e)
		return ErrFsInternal
	}
	logger.Info("[File] Store a file.",
		"dst", dst,
		"size", n,
	)
	return nil
}

func digestToHex(b []byte) string {
	h := sha256.New()
	// add datetime string to avoid conflict
	h.Write([]byte(time.Now().Format(time.RFC3339)))
	h.Write(b)
	return hex.EncodeToString(h.Sum(nil))
}
