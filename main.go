package main

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var (
	countHasDosBox int
	countNotDosBox int
	filesUpdated   = make([]string, 0)
	aspectRegex    = regexp.MustCompile("\naspect=false")
)

func find(root, ext string) []string {
	var a []string
	filepath.WalkDir(root, func(s string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if filepath.Ext(d.Name()) == ext {
			a = append(a, s)
		}
		return nil
	})
	return a
}

func main() {
	// default steamapps common dir: ~/.local/share/Steam/steamapps/common
	// example steamapps common dir on steam deck sd card: /run/media/mmcblk0p1/steamapps/common
	var dir = "~/.local/share/Steam/steamapps/common"
	if len(os.Args) >= 2 {
		dir = os.Args[1]
	}

	// attempt to update all files that end in confg
	for _, s := range find(dir, ".conf") {
		err := fixAspectRatio(s)
		if err != nil {
			log.Fatalf("failed to update aspect ratio: %s", err.Error())
		}
	}

	log.Printf("countHasDosBox: %d\n", countHasDosBox)
	log.Printf("countNotDosBox: %d\n", countNotDosBox)

	filesUpdatedMessage := fmt.Sprintf("files updated: %d\n%v", len(filesUpdated), strings.Join(filesUpdated, "\n"))
	log.Println(filesUpdatedMessage)

	f, err := os.Create(fmt.Sprintf("./files-updated-%s.txt", time.Now().Format(time.RFC3339)))
	if err != nil {
		log.Fatalf("failed to create file: %s", err.Error())
	}
	defer f.Close()

	_, err = f.Write([]byte(filesUpdatedMessage))
	if err != nil {
		log.Fatalf("failed to write to file: %s", err.Error())
	}
}

// fixAspectRatio will update the dosbox aspect config from false to true for all files that contain the word dosbox
// there's probably better ways to ensure that we're only updating dosbox conf files though
func fixAspectRatio(path string) error {
	read, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file: %s - %w", path, err)
	}
	if strings.Contains(strings.ToLower(string(read)), "dosbox") {
		countHasDosBox++
	} else {
		countNotDosBox++
		return nil
	}
	if aspectRegex.MatchString(string(read)) {
		filesUpdated = append(filesUpdated, path)
	}
	log.Printf("updating file: %s", path)
	newContents := aspectRegex.ReplaceAllString(string(read), "\naspect=true")
	err = ioutil.WriteFile(path, []byte(newContents), 0)
	if err != nil {
		return fmt.Errorf("failed to write to file: %s - %w", path, err)
	}
	return nil
}
