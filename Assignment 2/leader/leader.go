package main

import (
	"fmt"
	"io"
	"net/rpc"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type Args struct{ Words string }

func main() {

	var words map[string]int = make(map[string]int)

	//Get command-line argument for folder directory
	cliArg := os.Args[1]
	files, err := FilePathWalkDir(cliArg)
	if err != nil {
		fmt.Println("File Error")
		panic(err)
	}
	
	wg := new(sync.WaitGroup)
	mu := new(sync.Mutex)
	client, err := rpc.Dial("tcp", "localhost:12019")

	//For each file...
	for i := 0; i < len(files); i++ {
		matchString := ""
		fmt.Printf("Reading File: %s...\n", files[i])
		//Open new file
		file, _ := os.Open(files[i])
		defer file.Close()

		var headPos int64 = 0
		fMeta, _ := os.Stat(files[i])

		for headPos < fMeta.Size() {
			
			//Read 100 bytes from file
			buffer := make([]byte, min(100, fMeta.Size() - headPos))
			_, _ = file.Read(buffer)

			//Convert to string
			snip := string(buffer[:])

			//Find last index of space
			cut := strings.LastIndex(snip, " ")

			//cut string to end of last word in string
			if headPos + 100 < fMeta.Size() {
				matchString = snip[:cut + 1]
			} else {
				matchString = snip
			}

			/////////////////// RPC ///////////////////
			args := Args{Words: matchString}
			wg.Add(1)
			go callWordcount(&args, err, client, wg, words, mu)

			/////////////////////////////////////////////////////////////////////////////////////

			//get length of string
			bufLength := len(matchString)

			//Move read pointer bufLength byes further in file
			_, _ = file.Seek(headPos+int64(bufLength), io.SeekStart)
			headPos += int64(bufLength)
		}
	}

	wg.Wait()

	fmt.Println(words)

	
	// writeToFile(words, "output/outputFile")
}

func callWordcount(args *Args, err error, client *rpc.Client, wg *sync.WaitGroup, outputMap map[string]int, mu *sync.Mutex)   {
	defer wg.Done()

	var result map[string]int
	err = client.Call("WordCounter.CountWords", args, &result)

	if err != nil {
		fmt.Println("Error calling ", err)
		return
	}

	mu.Lock()
	for w, _ := range result {
		outputMap[w] += result[w]
	}
	mu.Unlock()

	return
	
}

func FilePathWalkDir(root string) ([]string, error) {
	var files []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func writeToFile(words map[string]int, fileName string) {
	outputFile, err := os.Create(fileName + ".txt")
	if err != nil {
		fmt.Println(err)
	}
	defer outputFile.Close()

	for word, freq := range words {
		fmt.Fprintln(outputFile, word, freq)
	}

	fmt.Println("Saved to output/" + fileName)
}
