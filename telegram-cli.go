package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aliforever/go-tdlib"
	"github.com/aliforever/go-tdlib/config"
	"github.com/aliforever/go-tdlib/entities"
	"github.com/aliforever/go-tdlib/incomingevents"
	"github.com/aliforever/go-tdlib/outgoingevents"
)

type Client struct {
	client      *tdlib.TDLib
	title       string
	configDir   string
	apiId       int64
	apiHash     string
	phoneNumber string
	chatId      int64
	message     string
	message_id  int64
	log         string
}

func main() {
	if strings.Contains(os.Args[1], "help") {
		fmt.Printf("usage. \n telegram-send configDir=./tg apiId=123 apiHash=xxx chatId=xxx log=file message=bla bla \n send only telegram-send message=bla bla \n")
		return
	}

	c := NewClient("sender")
	c.setVariables()
	if err := c.Start(); err != nil {
		exitProgram(fmt.Errorf("client not started error(%v)", err))
	}

}

func (c *Client) setVariables() {

	for _, v := range os.Args[:] {
		if strings.Contains(v, "apiId=") {
			i, err := strconv.ParseInt(strings.TrimLeft(v, "apiId="), 10, 64)
			if err != nil {
				exitProgram(fmt.Errorf("apiId invalid (%v)", err))
			}
			c.apiId = i
			continue
		}
		if strings.Contains(v, "chatId=") {
			i, err := strconv.ParseInt(strings.TrimLeft(v, "chatId="), 10, 64)
			if err != nil {
				exitProgram(fmt.Errorf("chatId invalid (%v)", err))
			}
			c.chatId = int64(i)
			continue
		}
		if strings.Contains(v, "apiHash=") {
			c.apiHash = strings.TrimPrefix(v, "apiHash=")
			continue
		}
		if strings.Contains(v, "message=") {
			c.message = strings.TrimPrefix(v, "message=")
		}
		if strings.Contains(v, "configDir=") {
			c.configDir = strings.TrimPrefix(v, "configDir=")
		}
		if strings.Contains(v, "log=") {
			c.configDir = strings.TrimPrefix(v, "log=")
		}
	}
	if c.configDir == "" {
		c.configDir = "./tdlib-db"
	}
	if c.log == "" {
		c.log = "/dev/null"
	}
	if c.apiId == 0 {
		c.apiId = 1
	}
	if c.apiHash == "" {
		c.apiHash = " "
	}

}

func NewClient(title string) *Client {
	return &Client{title: title}
}

func (c *Client) Start() error {
	cfg := config.New().
		SetDatabaseDirectory(c.configDir).
		IgnoreFileNames()

	h := tdlib.NewHandlers().
		SetOnUpdateAuthorizationStateEventHandler(c.onAuthorizationStateChange).
		// SetRawIncomingEventHandler(c.onRawIncomingEvent).
		// AddOnNewMessageHandler(c.onNewOutgoingMessage, tdlib.NewMessageFilters().SetIsOutgoingTrue()).
		SetErrorHandler(c.onError)

	manager := tdlib.NewManager(nil, tdlib.NewManagerOptions().SetLogPath(c.log))

	c.client = manager.NewClient(c.apiId, c.apiHash, h, cfg, nil)

	return c.client.ReceiveUpdates()
}
func (c *Client) onError(errr incomingevents.ErrorEvent) {
	fmt.Printf("Error: %v\n", errr)
}

func (c *Client) onAuthorizationStateChange(newState entities.AuthorizationStateType) {
	switch newState {
	case entities.AuthorizationStateTypeAwaitingTdlibParameters:
		err := c.client.SetTdlibParameters()
		if err != nil {
			fmt.Printf("Error: %s\n", err.Error())
			return
		}
	case entities.AuthorizationStateTypeAwaitingPhoneNumber:
		c.receivePhoneNumber()
	case entities.AuthorizationStateTypeAwaitingCode:
		c.receiveCode()
	case entities.AuthorizationStateTypeAwaitingPassword:
		c.receivePassword()
	case entities.AuthorizationStateTypeAwaitingRegistration:
		err := c.client.RegisterUser("Ley", "Johnson")
		if err != nil {
			fmt.Printf("Error: %s\n", err.Error())
			return
		}
	case entities.AuthorizationStateTypeReady:
		me, err := c.client.GetMe()
		if err != nil {
			exitProgram(err)
		}
		fmt.Printf("client %s %s connected. id(%v)\n", me.FirstName, me.LastName, me.Id)

		err = c.client.LoadChats(nil, 10)
		if err != nil {
			exitProgram(err)
		}

		if c.message == "" || c.chatId == 0 {
			fmt.Printf("Message not sended chatId(%v) message(%v)", c.chatId, c.message)
			exitProgram(err)
		}
		result, err := c.client.SendMessage(
			c.chatId,
			0,
			nil,
			&entities.InputMessageText{Text: &entities.FormattedText{Text: c.message}},
			&outgoingevents.SendMessageOptions{})
		if err != nil {
			fmt.Printf("message not send. error(%v)", err)
			exitProgram(err)
		}
		c.message_id = result.Id
		time.Sleep(200 * time.Millisecond)
		exitProgram(nil)
	default:
		fmt.Printf("Unhandled Authorization State: %s\n", newState)
	}
}

func (c *Client) receivePhoneNumber() {
	if c.phoneNumber == "" {
		fmt.Print("Phone Number:")
		fmt.Scan(&c.phoneNumber)
	}
	err := c.client.SetAuthenticationPhoneNumber(c.phoneNumber, &entities.PhoneNumberAuthenticationSettings{AllowFlashCall: false})
	if err != nil {
		fmt.Printf("Error: %v\n", err.Error())
	}
}

func (c *Client) receiveCode() {
	fmt.Print("Enter Code: ")
	var code string
	fmt.Scan(&code)
	err := c.client.CheckAuthenticationCode(code)
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
	}
}

func (c *Client) receivePassword() {
	fmt.Print("Enter Password: ")
	var password string
	fmt.Scan(&password)
	err := c.client.CheckAuthenticationPassword(password)
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
	}
}

// func (c *Client) onNewOutgoingMessage(message *incomingevents.UpdateNewMessage) {
// }

// func (c *Client) onRawIncomingEvent(bytes []byte) {
// 	// return
// }

func exitProgram(err error) {
	if err == nil {
		os.Exit(0)
	}
	os.Exit(255)

}
