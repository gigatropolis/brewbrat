package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"../brewbrat/calc"
	"../brewbrat/ingredients"
	"github.com/nlopes/slack"
)

var botID string = "<@" + os.Getenv("BOT_ID") + ">"
var botAlias = "@"

// Errors is truct for errors interface
// num is error number
// Message error string
type Errors struct {
	Num     uint64
	Message string
}

func (e *Errors) Error() string {
	return e.Message
}

// MesChannel is mai channel all events channel through
var MesChannel *chan slack.RTMEvent // := make(chan slack.RTMEvent)

// Connecter for handling event messages
type Connecter interface {
	GetMessageChannel() *chan slack.RTMEvent
	SendMessage(Message, Channel string)
}

// SlackConnector connection interface to Slack channel using nslopes
type SlackConnector struct {
	api *slack.Client
	rtm *slack.RTM
}

// StdInputConnector connection interface for command line standard input. used for test
type StdInputConnector struct {
	ChnIn chan slack.RTMEvent
}

func (s *StdInputConnector) scanner(in *os.File) {
	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
		msg := slack.RTMEvent{
			Type: "message",
			Data: &slack.MessageEvent{},
		}

		sMes := strings.Replace(scanner.Text(), botAlias, botID, 1)
		msg.Data.(*slack.MessageEvent).Channel = "My Channel"
		msg.Data.(*slack.MessageEvent).User = "Me"
		msg.Data.(*slack.MessageEvent).Text = sMes

		s.ChnIn <- msg
	}
}

// GetMessageChannel is connector interface method used to setup message channel
func (s *StdInputConnector) GetMessageChannel() *chan slack.RTMEvent {

	go s.scanner(os.Stdin)

	s.ChnIn = make(chan slack.RTMEvent)
	return &s.ChnIn
}

// SendMessage is connector interface method used to send message to output device
func (s *StdInputConnector) SendMessage(Message, Channel string) {
	fmt.Println(Message)
}

// GetMessageChannel is connector interface method used to setup message channel
// SlackConnector creates connection to Slack
func (s *SlackConnector) GetMessageChannel() *chan slack.RTMEvent {
	s.api = slack.New(
		os.Getenv("SLACK_BOT_TOKEN"),
		slack.OptionDebug(true),
		slack.OptionLog(log.New(os.Stdout, "slack-bot: ", log.Lshortfile|log.LstdFlags)),
	)

	s.rtm = s.api.NewRTM()
	go s.rtm.ManageConnection()

	return &s.rtm.IncomingEvents

}

// SendMessage is connector interface method used to send message to output device
func (s *SlackConnector) SendMessage(Message, Channel string) {
	s.rtm.SendMessage(s.rtm.NewOutgoingMessage(Message, Channel))
}

// GetHelpMessage reurns full help message read from README.md
func GetHelpMessage() string {
	dat, err := ioutil.ReadFile("README.md")
	if err != nil {
		return "error reading help file: " + err.Error()
	}

	return string(dat)
}

// HandleCommand main function call to handle messages to brewbrat
// and call correct APIs
func HandleCommand(message string) (string, error) {
	var err error
	words := strings.Split(strings.TrimSpace(message), " ")
	cmdResponse := ""
	for c, w := range words {
		words[c] = strings.TrimSpace(strings.ToLower(w))
		fmt.Printf("words[%d] = '%s'\n", c, w)
	}

	switch {
	case words[0] == "help":

		cmdResponse = GetHelpMessage()

	case words[0] == "calc":
		cmdResponse, err = calc.HandleCommand(words, message)
		if err != nil {
			fmt.Printf("calc returned error::%s\n", err.Error())
		}
	case words[0] == "list" || words[0] == "ls":
		cmdResponse, err = ingredients.HandleList(words, message)
	case words[0] == "explain" || words[0] == "ex":
		cmdResponse, err = ingredients.HandleExplanation(words, message)

	}

	return cmdResponse, err
}

// HandleMessageEvent is main handler to event messages
func HandleMessageEvent(ev *slack.MessageEvent) (string, error) {
	fmt.Printf("\nev.Text=%s\n", ev.Text)
	iStart := strings.Index(ev.Text, botID)
	if iStart < 0 {
		return "", &Errors{1, "NotForMe"}
	}
	fmt.Printf("\niStart = %d\n", iStart)
	msg := ev.Text
	iStart += len(botID)
	if iStart > len(msg) {
		return "", &Errors{2, "Index out of range"}
	}

	msg = msg[iStart:]
	retMsg, err := HandleCommand(msg)
	if err != nil {
		return "Cannot handle event text: " + err.Error(), err
	}

	return retMsg, nil // fmt.Sprintf("<%s> '%s'\n", ev.Channel, retMsg), nil
}

func main() {

	ingredients.GetBeerXMLFromFile("")

	var conn Connecter

	if len(os.Args) > 1 && os.Args[1] == "test" {
		conn = &StdInputConnector{}
	} else {
		conn = &SlackConnector{}
	}

	MesChannel = conn.GetMessageChannel()

	for msg := range *MesChannel { // rtm.IncomingEvents
		fmt.Print("Event Received: ")
		switch ev := msg.Data.(type) {
		case *slack.HelloEvent:
			// Ignore hello

		case *slack.ConnectedEvent:
			fmt.Println("Infos:", ev.Info)
			fmt.Println("Connection counter:", ev.ConnectionCount)
			// Replace C2147483705 with your Channel ID
			// conn.SendMessage("Hello world", "C496VLJ2D")

		case *slack.MessageEvent:
			fmt.Printf("Message: %v\n", ev)
			message, err := HandleMessageEvent(ev)
			if err != nil {
				if err.Error() == "NotForMe" {
					continue
				} else {
					conn.SendMessage(fmt.Sprintf("Error parsing message Event: %s", err.Error()), ev.Channel)
				}
			} else {
				conn.SendMessage(message, ev.Channel)
			}
		case *slack.PresenceChangeEvent:
			fmt.Printf("Presence Change: %v\n", ev)

		case *slack.LatencyReport:
			fmt.Printf("Current latency: %v\n", ev.Value)

		case *slack.RTMError:
			fmt.Printf("Error: %s\n", ev.Error())

		case *slack.InvalidAuthEvent:
			fmt.Printf("Invalid credentials")
			return

		default:

			// Ignore other events..
			fmt.Printf("Unexpected: %v\n", msg.Data)
		}
	}
}
