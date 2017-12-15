package main

import (
	"gopkg.in/natefinch/lumberjack.v2" //rotational logging
	"fmt"
	"log"
	"net"
	"regexp"
	"github.com/ClinicalSystemsEngineering/tap"
	//"strconv"
	"io"
	"sip2tap/sipparser"
	"strings"
	//"time"
	"flag"
)

func handleUDPConnection(c *net.UDPConn, parsedmsgs chan<- string) {

	//fmt.Print("Listening...\n")

	var message string //concatenated SIP message
	var firstline string
	var bytes = make([]byte, 1536) //connection reading buffer
	var bytesread int
	var byteswritten int
	var trycntr int
	//read the SIP data and process it
	//fmt.Print("Reading...\n\n")
	message = ""

TryReadingAgain:
	bytesread, raddr, err := c.ReadFromUDP(bytes)
	//readstring, err := r.ReadString('\n')
	switch {

	case err == io.EOF:
		log.Println("Reached EOF.\n   ---")
		return
	case err != nil:
		log.Printf("\nError reading SIP data: %v'", err)
		return
	case bytesread == 0:
		if trycntr < 3 {
			log.Println("Zero bytes read in message trying again.")
			trycntr++
			goto TryReadingAgain
		} else {
			log.Printf("Tried reading %v times. Giving up...", trycntr)
			return
		}
	}

	message = message + string(bytes)
	firstline = message[:strings.Index(message, "\n")+1]
	//fmt.Printf("processed first line:<%v>\n", strconv.QuoteToASCII(firstline))
	//catch ACK/NAK/BYE messages and process

	if strings.Contains(firstline, "ACK") || strings.Contains(firstline, "NAK") || strings.Contains(firstline, "BYE") {
		//fmt.Printf("Received ACK/NAK/BYE closing connection...\n")
		//c.Close()
		return

	}

	//send the message to a SIP parser to parse it and add it to a processed queue for tap output
	go sipparser.Parse(parsedmsgs, message)
	//find the end of the first message line and the end of message
	messageend := strings.Index(message, "\r\n\r\n") + 4
	//build CANCEL response
	response := "SIP/2.0 487 Request Terminated\r\n"
	response = response + message[len(firstline):messageend]
	//change content length to 0 and replace INVITE with CANCEL
	contentre := regexp.MustCompile(`Content-Length:\s.*\r\n`)
	response = contentre.ReplaceAllString(response, "Content-Length: 0\r\n")
	//response = strings.Replace(response,"INVITE","CANCEL",2)
	//fmt.Printf("Responding\n<%v>\n", strconv.QuoteToASCII(response))
	//write the CANCEL response back to the caller
	trycntr = 0
TryWritingAgain:
	byteswritten, err = c.WriteToUDP([]byte(response), raddr)
	switch {
	case err != nil:
		log.Printf("\nError writing SIP response %v\n", err)
		//c.Close()
		return
	case byteswritten == 0:
		if trycntr < 3 {
			log.Println("Zero bytes written in response trying again")
			trycntr++
			goto TryWritingAgain
		} else {
			log.Printf("Tried writing %v times. Giving up...\n", trycntr)
			return
		}
	}

	return
}

//main can accept 2 flag arguments the port for the SIP listener and the port
//for the TAP output listener
//i.e call sip2tap -sipPort=5080 tapPort=10001
//default ports are 5080 for SIP and 10001 for TAP

func main() {
	sipPort := flag.String("sipPort","5080","SIP listener port on local host")
	tapPort := flag.String("tapPort","10001","TAP listener port on local host")
	flag.Parse()
	log.SetOutput(&lumberjack.Logger{
		Filename:   "/var/log/sip2tap/sip2tap.log",
		MaxSize:    100, // megabytes
		MaxBackups: 5,
		MaxAge:     60,   //days
		Compress:   true, // disabled by default
	})
	
	log.Printf("STARTING SIP Listener on port udp %v...\n\n",*sipPort)
	// Listen on udp port 5080 on all available unicast and
	// anycast IP addresses of the local system.
	addr, _ := net.ResolveUDPAddr("udp", ":"+ *sipPort)

	sip, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Println("Error opening sip listener failing application.")
		log.Fatal(err)
	}
	defer sip.Close()

	//message processing channel for sip2tap conversions
	parsedmsgs := make(chan string, 1000)

	//start a tap server
	go tap.Server(parsedmsgs,*tapPort)

	for {
		// Handle UDP connections
		handleUDPConnection(sip, parsedmsgs)

	}
}
