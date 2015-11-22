//TODO handle channel closure
package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"strings"
	"time"
)

const CONNECTION_TIMEOUT = 4
const SPLIT_CHAR = "\n"

func handleClient(client net.Conn, flagsChannel chan string, logChannel chan string) {
	logChannel <- "HANDLER: serving client @" + client.RemoteAddr().String()
	client.SetReadDeadline(time.Now().Add(4 * time.Second))
	defer client.Close()
	input, err := ioutil.ReadAll(client)
	if err != nil && err != io.EOF {
		logChannel <- "HANDLER: " + err.Error()
	}
	tosend := string(input)
	logChannel <- "HANDLER: got flags:" + tosend
	flagsChannel <- tosend
}

func serve(flagsChannel chan string, logChannel chan string) {
	logChannel <- "SERVE: starting listener"
	listener, err := net.Listen("tcp", ":31337")
	if err != nil {
		logChannel <- "SERVE: unable to start listener"
		close(logChannel)
		return
	}
	defer listener.Close()
	logChannel <- "SERVE: started listener"
	for {
		client, err := listener.Accept()
		logChannel <- "SERVE: got client"
		if err != nil {
			logChannel <- "SERVE: error while handling client: " + err.Error()
			continue
		}
		go handleClient(client, flagsChannel, logChannel)
	}
}

func filter(flagsChannel chan string, filteredFlagsChannel chan string, logChannel chan string) {
	submittedFlags := make(map[string]bool)
	for flagsBulk, ok := <-flagsChannel; ok; flagsBulk, ok = <-flagsChannel {
		flags := strings.Split(flagsBulk, SPLIT_CHAR)
		for _, flag := range flags {
			if len(flag) > 0 && !submittedFlags[flag] {
				submittedFlags[flag] = true
				filteredFlagsChannel <- flag
			}
		}
	}
}

func submit(filteredFlagsChannel chan string, logChannel chan string) {
	delay := time.Millisecond * 500
	for flag, ok := <-filteredFlagsChannel; ok; flag, ok = <-filteredFlagsChannel {
		//TODO actual implementation changes for every competition
		conn, err := net.Dial("tcp", "localhost:31338")
		if err != nil {
			logChannel <- "SUBMIT: " + err.Error()
			logChannel <- "SUBMIT: retrying in " + delay.String()
			filteredFlagsChannel <- flag
			time.Sleep(delay)
			delay = delay * 2
			if delay.Seconds() > 30 {
				delay = 30 * time.Second
			}
			continue
		} else {
			delay = time.Millisecond * 500
		}
		fmt.Fprintf(conn, flag)
		status, err := bufio.NewReader(conn).ReadString('\n')
		//TODO handle statuses
		logChannel <- "SUBMIT: submitted: " + flag + " status: " + status
		conn.Close()
	}

}

func logger(logChannel chan string) {
	for logLine, ok := <-logChannel; ok; logLine, ok = <-logChannel {
		fmt.Println(logLine)
	}
	fmt.Println("Something went wrong, terminating...")
}

func main() {
	flagsChannel := make(chan string, 4096)
	filteredFlagsChannel := make(chan string, 4096)
	logChannel := make(chan string, 4096)
	logChannel <- "STARTING SUBMITTER"
	go serve(flagsChannel, logChannel)
	go filter(flagsChannel, filteredFlagsChannel, logChannel)
	go submit(filteredFlagsChannel, logChannel)
	logger(logChannel)
}
