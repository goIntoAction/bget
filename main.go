package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	osPath "path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gammazero/workerpool"
)

var files []os.FileInfo
var entryFileName = "entry.json"
var pool *workerpool.WorkerPool

var wg sync.WaitGroup

func main() {
	var sourceDir, outDir string
	flag.StringVar(&sourceDir, "source", "", "source dir")
	flag.StringVar(&outDir, "out", "./", "out dir")
	flag.Parse()

	if sourceDir == "" {
		return
	}
	pool = workerpool.New(5)
	if !strings.HasSuffix(outDir, "/") {
		outDir = outDir + "/"
	}
	handleFiles(sourceDir, outDir)
	pool.StopWait()
	log.Print("done")
}

func handleFiles(root, outDir string) {
	if exists, _ := pathExists(outDir); !exists {
		os.MkdirAll(outDir, os.ModePerm)
	}
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && strings.Compare(info.Name(), entryFileName) == 0 {
			data, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}
			str := string(data)
			if strings.HasSuffix(str, "}}}}") {
				data = data[:len(data)-2]
			} else if strings.HasSuffix(str, "}}}") {
				data = data[:len(data)-1]
			}
			entry := &entry{}
			err = json.Unmarshal(data, entry)
			if err != nil {
				log.Print(err)
				return err
			}
			parent := osPath.Dir(path)
			fileInfos, err := ioutil.ReadDir(parent)
			if err != nil {
				return err
			}
			for _, dataDir := range fileInfos {
				if dataDir.IsDir() {
					audioFile := osPath.Join(osPath.Dir(path), dataDir.Name(), "audio.m4s")
					videoFile := osPath.Join(osPath.Dir(path), dataDir.Name(), "video.m4s")
					cmd := fmt.Sprintf("ffmpeg -i '%s' -i '%s' -c:v copy -c:a aac -strict experimental '%s'", videoFile, audioFile, outDir+entry.pageData.part+".mp4")
					pool.Submit(func() {
						command(cmd)
					})
					break
				}
			}
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	for _, fileInfo := range files {
		fileInfo.Name()
	}
}

func command(command string) (string, error) {
	cmd := exec.Command("bash", "-c", command)
	out, err := cmd.CombinedOutput()
	result := string(out)
	return result, err
}

type entry struct {
	pageData pageData `json:"page_data"`
	title    string   `json:"title`
}

type pageData struct {
	part string `json:"part"`
}

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
