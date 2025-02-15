package main

import (
	"fmt"
	"io"
	"net/rpc"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sort"
	// "regexp"
	// "time"
)

type Args struct{ Words string }

func main() {

	var words map[string]int = make(map[string]int)
	numClients := 2
	totalNumMessages := 0
	var ipAddrs = [2]string{"localhost:12019", "localhost:12020"}
	var clients [2]*rpc.Client

	//Get command-line argument for folder directory
	cliArg := os.Args[1]
	files, err := FilePathWalkDir(cliArg)
	if err != nil {
		fmt.Println("File Error")
		panic(err)
	}
	
	wg := new(sync.WaitGroup)
	mu := new(sync.Mutex)
	for i := 0; i < numClients; i++ {
		clients[i], _ = rpc.Dial("tcp", ipAddrs[i])
	}

	//For each file...
	for i := 0; i < len(files); i++ {
		if strings.Contains(string(files[i]), ".txt") {
			matchString := ""



			fmt.Printf("Reading File: %s...\n", files[i])
			//Open new file
			file, _ := os.Open(files[i])
			defer file.Close()

			var headPos int64 = 0
			fMeta, _ := os.Stat(files[i])

			var outRequests int64 = 0;
			fmt.Println(fMeta.Size());
			// time.Sleep(10000 * time.Second)


			for headPos < fMeta.Size() {
				if outRequests < 2 {
					//Read 100 bytes from file
					buffer := make([]byte, min(100, fMeta.Size() - headPos))
					_, _ = file.Read(buffer)

					//Convert to string
					snip := string(buffer[:])

					// snip = regexp.MustCompile(`\s`).ReplaceAllString(snip, " ")

					//Find last index of space
					cut := strings.LastIndex(snip, " ")

					fmt.Println(snip)

					// time.Sleep(time.Second / 10)

					//cut string to end of last word in string
					if headPos + 100 < fMeta.Size() && cut > 0 {
						matchString = snip[:cut]
					} else {
						matchString = snip
						cut = int(min(100, fMeta.Size() - headPos))
					}



					/////////////////// RPC ///////////////////
					args := Args{Words: matchString}
					wg.Add(1)

					go callWordcount(&args, err, clients[totalNumMessages % numClients], wg, words, mu, &outRequests)
					outRequests++
					totalNumMessages++
					fmt.Println(headPos)

					/////////////////////////////////////////////////////////////////////////////////////

					//Move read pointer bufLength byes further in file
					_, _ = file.Seek(headPos+int64(cut), io.SeekStart)
					headPos += int64(cut)
				}
			}
		}
	}

	wg.Wait()

	fmt.Println(words)

	//Alphabetize list
	keys := make([]string, 0, len(words))
	for key := range words {
        keys = append(keys, key)
    }
	sort.Strings(keys)

	f, _ := os.Create("./output/results.txt")
	defer f.Close()


	var totalBytes int = 0

	for _, key := range keys {
		n_bytes, _ := f.WriteString(fmt.Sprintf("%s %d\n", string(key), words[key]))
		totalBytes += n_bytes
	}

	fmt.Printf("Sorted %d words\n", len(keys))

	f.Sync()

    fmt.Printf("wrote %d bytes to file \n", totalBytes)

	
	// writeToFile(words, "output/outputFile")
}

func callWordcount(args *Args, err error, client *rpc.Client, wg *sync.WaitGroup, outputMap map[string]int, mu *sync.Mutex, outRequests *int64)   {
	defer wg.Done()

	var result map[string]int
	err = client.Call("WordCounter.CountWords", args, &result)
	// fmt.Println("asdasd")

	if err != nil {
		fmt.Println("Error calling ", err)
		return
	}

	mu.Lock()
	for w, _ := range result {
		outputMap[w] += result[w]
	}

	*outRequests--
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
