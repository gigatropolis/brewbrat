package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/nlopes/slack"
)

type Errors struct {
	Num     uint64
	Message string
}

func (e *Errors) Error() string {
	return fmt.Sprintf("%5d: ", e.Num) + e.Message
}

var BOT_ID string = "<@" + os.Getenv("BOT_ID") + ">"

func HandleMessage(ev *slack.MessageEvent) (string, error) {
	fmt.Printf("\nev.Text=%s\n", ev.Text)
	iStart := strings.Index(ev.Text, BOT_ID)
	if iStart < 0 {
		return "", &Errors{1, "Could not find starting index"}
	}
	fmt.Printf("\niStart = %d\n", iStart)
	msg := ev.Text
	iStart += 13
	if iStart > len(msg) {
		return "", &Errors{2, "Index out of range"}
	}

	msg = msg[iStart:]
	retMsg := fmt.Sprintf("text: '%s', channel %s\n", msg, ev.Channel)
	return retMsg, nil
}

func main() {
	api := slack.New(
		os.Getenv("SLACK_BOT_TOKEN"),
		slack.OptionDebug(true),
		slack.OptionLog(log.New(os.Stdout, "slack-bot: ", log.Lshortfile|log.LstdFlags)),
	)

	rtm := api.NewRTM()
	go rtm.ManageConnection()

	for msg := range rtm.IncomingEvents {
		fmt.Print("Event Received: ")
		switch ev := msg.Data.(type) {
		case *slack.HelloEvent:
			// Ignore hello

		case *slack.ConnectedEvent:
			fmt.Println("Infos:", ev.Info)
			fmt.Println("Connection counter:", ev.ConnectionCount)
			// Replace C2147483705 with your Channel ID
			rtm.SendMessage(rtm.NewOutgoingMessage("Hello world", "C496VLJ2D"))

		case *slack.MessageEvent:
			fmt.Printf("Message: %v\n", ev)
			message, err := HandleMessage(ev)
			if err != nil {
				rtm.SendMessage(rtm.NewOutgoingMessage(fmt.Sprintf("Error parsing message: %s", err.Error()), ev.Channel))
			} else {
				rtm.SendMessage(rtm.NewOutgoingMessage(message, ev.Channel))
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
