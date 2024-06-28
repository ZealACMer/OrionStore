package objects

import (
	"crypto/sha256"
	"encoding/base64"
	"log"
	"net/url"
	"orionStore/dataServer/locate"
	"os"
	"path/filepath"
	"strings"
)

// getFile 根据散列值，查找对象文件
func getFile(name string) string {
	files, _ := filepath.Glob(os.Getenv("STORAGE_ROOT") + "/objects/" + name + ".*")
	if len(files) != 1 {
		return ""
	}
	file := files[0]
	h := sha256.New()
	sendFile(h, file)
	d := url.PathEscape(base64.StdEncoding.EncodeToString(h.Sum(nil)))
	hash := strings.Split(file, ".")[2]
	if d != hash {
		log.Println("object hash mismatch, remove", file)
		locate.Del(hash)
		os.Remove(file)
		return ""
	}
	return file
}
