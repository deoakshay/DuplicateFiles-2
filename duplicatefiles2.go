package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/stacktic/dropbox"
)

type DuplicateFiles interface {
	ListDirectories(string) []string
	HashAndWrite(string, *sync.Map, *sync.WaitGroup)
}
type OSLevel struct {
}
type DropBoxLevel struct {
	Clientid     string
	Clientsecret string
	TokenId      string
}

func (op OSLevel) HashAndWrite(path string, hashValueMap *sync.Map, wg *sync.WaitGroup) {
	defer wg.Done()
	fileNames, err := ioutil.ReadDir(path) //should use interface for ioutil.ReadDir
	if err != nil {
		log.Fatal("exiting in readdir", err)
	}
	for _, files := range fileNames {
		if !files.IsDir() {

			absolutePath := filepath.Join(path, files.Name())
			file, err := os.Open(absolutePath)
			if err != nil {
				log.Fatal("exiting in opening", err)
			}
			hashValue := md5.New()

			if _, err := io.Copy(hashValue, file); err != nil {

				log.Fatal("exiting in copy", err)
			}

			stringValueOfHash := hex.EncodeToString(hashValue.Sum(nil))
			if value, ok := hashValueMap.Load(stringValueOfHash); !ok {
				hashValueMap.Store(stringValueOfHash, []string{absolutePath})
			} else if fileArray, ok := value.([]string); ok {

				fileArray = append(fileArray, absolutePath)
				hashValueMap.Store(stringValueOfHash, fileArray)

			}
		}
	}

}

func (op OSLevel) ListDirectories(path string) []string {
	fmt.Println("In Local Computer:")
	fileNames, err := ioutil.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}
	directories := []string{}
	directories = append(directories, path)
	for _, files := range fileNames {

		if !files.IsDir() {
			continue
		} else {
			directories = append(directories, filepath.Join(path, files.Name()))
		}
	}
	return directories
}

func (dbl DropBoxLevel) ListDirectories(path string) []string {
	dropBoxObject := dropbox.NewDropbox()
	fmt.Println("In DropBox:")
	dropBoxObject.SetAccessToken(dbl.TokenId)
	dropBoxMetaData, _ := dropBoxObject.Metadata(path, true, true, "", "", 1000)
	directories := []string{}
	for index, _ := range dropBoxMetaData.Contents {
		if dropBoxMetaData.Contents[index].IsDir == true {
			directories = append(directories, dropBoxMetaData.Contents[index].Path)
		}

	}
	directories = append(directories, path)
	return directories
}
func (dbl DropBoxLevel) HashAndWrite(path string, hashValueMap *sync.Map, wg *sync.WaitGroup) {
	defer wg.Done()
	dropBoxObject := dropbox.NewDropbox()
	dropBoxObject.SetAccessToken(dbl.TokenId)
	dropBoxMetaData, _ := dropBoxObject.Metadata(path, true, true, "", "", 1000)
	for index, _ := range dropBoxMetaData.Contents {
		absolutePath := dropBoxMetaData.Contents[index].Path
		downloadedFile, size, _ := dropBoxObject.Download(absolutePath, "", 0)
		if size > 0 {
			hashValue := md5.New()
			if _, err := io.Copy(hashValue, downloadedFile); err != nil {

				log.Fatal("exiting in copy", err)
			}

			stringValueOfHash := hex.EncodeToString(hashValue.Sum(nil))
			if value, ok := hashValueMap.Load(stringValueOfHash); !ok {
				hashValueMap.Store(stringValueOfHash, []string{absolutePath})
			} else {
				fileArray, ok := value.([]string)
				if ok {
					fileArray = append(fileArray, absolutePath)
					hashValueMap.Store(stringValueOfHash, fileArray)
				}
			}

		}

	}
}

//write a free function traversing

func main() {
	var choice int
	fmt.Println("Do you want to run the program on:\n1.Local Files \n2.Dropbox Files:")
	fmt.Scan(&choice)
	var duplicateFilesObject DuplicateFiles
	var path string
	if choice == 1 {
		duplicateFilesObject = OSLevel{}
		path = "/Users/akshaydeo/Downloads"
	} else {
		duplicateFilesObject = DropBoxLevel{"31rmr26bffk3ij8", "n0rlqt27iuf7scp", "KeymFkX_8yAAAAAAAAACSBcyXS5BbpSCBxa4wf7ejZAdhEyt201sno3he5lImvl4"}
		path = "/"
	}
	dupes := findDuplicates(duplicateFilesObject, path)
	fmt.Println("The Duplicate Files are:")
	for _, value := range dupes {

		if len(value) > 1 {
			fmt.Println("\n", value)
		}
	}
}

func findDuplicates(d DuplicateFiles, path string) map[string][]string {
	hashValueMap := new(sync.Map)
	var wg sync.WaitGroup
	listDirectories := d.ListDirectories(path)
	fmt.Println(listDirectories)
	for directories := range listDirectories {
		wg.Add(1)
		go d.HashAndWrite(listDirectories[directories], hashValueMap, &wg)

	}
	wg.Wait()
	storeValues := make(map[string][]string)
	hashValueMap.Range(func(key, value interface{}) bool {
		storeValues[key.(string)] = value.([]string)
		return true
	})

	return storeValues
}
