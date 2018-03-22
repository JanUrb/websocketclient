package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/gorilla/websocket"
)

type config struct {
	connected   bool   //indicates the connection status. Always true for now.
	shortmode   bool   //indicates that the cli shows less or more information everytime the user inputs something
	host        string //contains the connected host. default value is "No host".
	path        string
	messageType int //indicates which message type is used for writing to the connection.
}

func (cfg *config) printLineStart() {
	if cfg.shortmode {
		if cfg.connected {
			fmt.Print("(connected):")
		} else {
			fmt.Print("(not connected):")
		}
	} else {
		if cfg.connected {
			fmt.Printf("(connected - %s%s - writemode: %s) :", cfg.host, cfg.path, messageTypeToString(cfg.messageType))
		} else {
			fmt.Print("(not connected):")
		}
	}
}

func messageTypeToString(messageType int) string {
	switch messageType {
	case websocket.BinaryMessage:
		return "BinaryMessage"
	case websocket.TextMessage:
		return "TextMessage"
	default:
		return "Not in use in this application"
	}
}

func main() {
	done := make(chan bool)

	host := "localhost:8080"
	path := "/ws"

	url := url.URL{Scheme: "ws", Host: host, Path: path}
	conn, _, err := websocket.DefaultDialer.Dial(url.String(), nil)

	cfg := &config{
		connected:   true,
		host:        host,
		path:        path,
		messageType: websocket.TextMessage,
		shortmode:   true,
	}

	if err != nil {
		log.Println("Errow while dialing: ", err)
		return
	}

	defer conn.Close()

	go writeConnection(done, cfg, conn)
	go readConnection(done, conn)
	<-done //wait until something goes wrong or the user exits the application
	log.Println("Good bye!")
}

func readConnection(done chan<- bool, conn *websocket.Conn) {
	defer func() { done <- true }()
	for {
		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error while reading message ", err)

			return
		}

		log.Println("Message type: ", msgType)
		log.Printf("Received message: %s", msg)
	}
}

func writeConnection(done chan<- bool, cfg *config, conn *websocket.Conn) {
	defer func() {
		done <- true
	}()
	cfg.printLineStart()

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		input := scanner.Text()

		if strings.HasPrefix(input, "from file") {
			trimmed := strings.TrimSpace(input)
			fp := trimmed[len("from file")+1:]
			fmt.Printf("sending from file %s\n", fp)
			b, err := ioutil.ReadFile(fp)
			if err != nil {
				log.Printf("Error reading file %s with error: %s\n", fp, err)
				continue
			}
			input = fmt.Sprintf("%s", b)
		}

		switch input {
		case "exit":
			return
		case "help":
			printHelp()
			cfg.printLineStart()
			continue
		case "host":
			fmt.Printf("host: %s\tpath: %s\n", cfg.host, cfg.path)
			cfg.printLineStart()
			continue
		case "mode text":
			cfg.messageType = websocket.TextMessage
			fmt.Print("Changed message type to text\n")
			cfg.printLineStart()
			continue
		case "mode binary":
			cfg.messageType = websocket.BinaryMessage
			fmt.Print("Changed message type to binary\n")
			cfg.printLineStart()
			continue
		case "mode":
			fmt.Printf("Message type is: %s\n", messageTypeToString(cfg.messageType))
			cfg.printLineStart()
			continue

		default:
			cfg.printLineStart()

			err := conn.WriteMessage(cfg.messageType, []byte(input))
			if err != nil {
				log.Println("Error while sending message: ", err)
				return
			}
		}

	}
}

func printHelp() {
	helpText := `
	exit: closes the application
	help: prints information about different commands
	host: prints information about the host the user is connected to
	mode: prints the currently used message type
	mode text: changes the message type to text message
	mode binary: changes the message type to binary message
	from file "filepath": reads the content of a file and sends it`

	fmt.Printf("%s\n\n", helpText)
}
