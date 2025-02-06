package main

import (
	"fmt"
	"net/rpc"
	"sync"
)

type Args struct{ Words string }

func main() {

	wg := new(sync.WaitGroup)

	client, err := rpc.Dial("tcp", "localhost:12019")
	// client2, err := rpc.Dial("tcp", "localhost:12020")

	for i := 0; i < 12; i++ {
		args := Args{Words: "1232 dasd asdfgh"}
		wg.Add(1)
		go callWordcount(&args, err, client, wg)
	}
	wg.Wait()
}

func callWordcount(args *Args, err error, client *rpc.Client, wg *sync.WaitGroup) {
	var result map[string]int
	err = client.Call("WordCounter.CountWords", args, &result)

	if err != nil {
		fmt.Println("Error calling ", err)
		return
	}
	fmt.Println(result)
	wg.Done()
}
