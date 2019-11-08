package sftpClient

import (
	"fmt"
	"github.com/pkg/sftp"
	"io/ioutil"
	"log"
)

// FindRemoteFiles enumerates files in a remote directory
func FindRemoteFiles(sftpClient *sftp.Client, path string) func() (r <-chan FileResponse) {
	return func() (r <-chan FileResponse) {
		responseChannel := make(chan FileResponse)

		go func() {
			defer close(responseChannel)
			stats, err := sftpClient.Lstat(path)
			if err != nil {
				responseChannel <- FileResponse{File: "", Err: fmt.Errorf("Cannot STAT 'remote:%s': %v\n", path, err)}
				return
			}
			if !stats.IsDir() {
				responseChannel <- FileResponse{File: "", Err: fmt.Errorf("'remote:%s' is not a directory", path)}
				return
			}

			var walker = *sftpClient.Walk(path)
			for walker.Step() {
				if err := walker.Err(); err != nil {
					responseChannel <- FileResponse{File: "", Err: err}
					continue
				}
				if walker.Path() == path {
					continue
				}
				responseChannel <- FileResponse{File: walker.Path(), Err: nil}
			}
		}()

		return responseChannel
	}
}

func findRemoteFilesAggregator(functions []func() (r <-chan FileResponse)) (r <-chan FileResponse) {
	responseChannel := make(chan FileResponse)

	go func(functions []func() (r <-chan FileResponse)) {
		for _, function := range functions {
			intermediateChannel := function()
			for response := range intermediateChannel {
				responseChannel <- response
			}
		}
		close(responseChannel)
	}(functions)

	return responseChannel
}

// FindAllRemoteFiles enumerates all remote files in multiple directories and their descendents
func FindAllRemoteFiles(sftpClient *sftp.Client, paths []string) ([]string, error) {
	var functions []func() (r <-chan FileResponse)
	var files []string

	for _, path := range paths {
		functions = append(functions, FindRemoteFiles(sftpClient, path))
	}

	responseChannel := findRemoteFilesAggregator(functions)
	encounteredErrors := 0
	for response := range responseChannel {
		if response.Err != nil {
			encounteredErrors++
			log.Println(response.Err)
		}
		if encounteredErrors == 0 {
			files = append(files, response.File)
		}
	}

	if encounteredErrors > 0 {
		return nil, fmt.Errorf("Encountered %d errors\n", encounteredErrors)
	}
	return files, nil
}

// FindAllLocalFiles enumerates all local files in multiple directories and their descendents
func FindAllLocalFiles(paths []string) ([]string, error) {
	filesArray := make([]string, 0)
	if len(paths) > 0 {
		for _, fileName := range paths {
			files, err := ioutil.ReadDir(fileName)
			if err != nil {
				log.Fatalf("ERROR: Could not read collection folder %s\r\n", fileName)
			}
			for _, file := range files {
				if file.IsDir() {
					continue //skip directories
				}
				filesArray = append(filesArray, file.Name())
			}
		}
	}
	return filesArray, nil
}
