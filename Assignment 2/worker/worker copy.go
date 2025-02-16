package main

import (
	"errors"
	"fmt"
	"net"
	"net/rpc"
	"regexp"
	"strings"
)

type WordCounter struct{}
type Args struct{ Words string }

func (w *WordCounter) CountWords(args *Args, reply *map[string]int) error {
	outputMap := make(map[string]int)
	if args == nil {
		return errors.New("Error")
	}

	line := strings.ToLower(args.Words)
	line = regexp.MustCompile(`[^a-zA-Z0-9 ]+`).ReplaceAllString(line, " ")
	words := strings.Fields(line)

	for _, word := range words {
		outputMap[word] += 1
	}

	*reply = outputMap
	return nil
}

func main() {
	//Create new RPC handler
	wc := new(WordCounter)
	//Register wordcounter to rpc service
	rpc.Register(wc)

	listener, err := net.Listen("tcp", "localhost:12020")
	if err != nil {
		fmt.Println("Error ", err)
		return
	}
	defer listener.Close()
	rpc.Accept(listener)
}
