package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type WordCountResult struct {
	WorkerId  uint16
	Filepath  string
	WordCount uint32
}

func main() {
	const rootPath = "texts"

	var files []string

	err := filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("Error occured during traversal of %s: %w", rootPath, err)
		}

		if d.IsDir() {
			return nil
		}

		if filepath.Ext(path) != ".txt" {
			return nil
		}

		files = append(files, path)

		// fmt.Printf("Found file: %s\n", path)

		return nil
	})
	if err != nil {
		fmt.Printf("Failed to traverse %s: %v\n", rootPath, err)
	}

	fmt.Printf("Found %d files\n", len(files))

	results := make(chan WordCountResult, len(files))

	var wg sync.WaitGroup

	fmt.Printf("Starting go routines\n")

	for i, file := range files {
		wg.Add(1)

		go countWords(uint16(i), file, results, &wg)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var wordSum uint32 = 0

	for result := range results {
		wordSum += result.WordCount

		fmt.Printf("Goroutine-worker-%d -> %s: %d słów\n", result.WorkerId, filepath.Base(result.Filepath), result.WordCount)
	}

	fmt.Printf("\nTotal word count: %d\n", wordSum)
}

func countWords(workerId uint16, filename string, results chan WordCountResult, wg *sync.WaitGroup) {
	defer wg.Done()

	content, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("Worker #%d failed to read file %s: %v\n", workerId, filename, err)
		results <- WordCountResult{WorkerId: workerId, Filepath: filename, WordCount: 0}
		return
	}

	text := string(content)

	wordCount := uint32(len(strings.Fields(text)))

	results <- WordCountResult{WorkerId: workerId, Filepath: filename, WordCount: wordCount}
}
