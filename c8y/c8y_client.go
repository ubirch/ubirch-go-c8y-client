package c8y

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)
import MQTT "github.com/eclipse/paho.mqtt.golang"

type c8yResponseForm struct {
	Password string `json:"password"`
	Tenant   string `json:"tenantId"`
	Self     string `json:"self"`
	ID       string `json:"id"`
	User     string `json:"username"`
}

func BootstrapHTTP(uuid string, tenant string, password string) (string, string) {
	data, err := json.Marshal(map[string]string{
		"id": uuid,
	})
	if err != nil {
		panic(err)
	}

	timeout := 5 * time.Second
	client := http.Client{
		Timeout: timeout,
	}

	request, err := http.NewRequest("POST",
		"https://ubirch.cumulocity.com/devicecontrol/deviceCredentials",
		bytes.NewBuffer(data))
	if err != nil {
		panic(err)
	}
	request.Header.Set("Content-Type", "application/vnd.com.nsn.cumulocity.deviceCredentials+json")
	request.Header.Set("Accept", "application/vnd.com.nsn.cumulocity.deviceCredentials+json")
	request.SetBasicAuth("devicebootstrap", password)

	for {
		resp, err := client.Do(request)
		if err != nil {
			log.Panic(err.Error())
		}
		//defer resp.Body.Close()

		log.Println("Response status:", resp.Status)

		if resp.StatusCode == http.StatusCreated {
			responseForm := c8yResponseForm{}

			// read response body
			bodyBytes, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Fatalf("ERROR: unable to read response: %v", err)
			}

			// parse json response body
			err = json.Unmarshal(bodyBytes, &responseForm)
			if err != nil {
				log.Fatalf("ERROR: unable to parse response: %v", err)
			}

			bodyString := string(bodyBytes)
			log.Println(bodyString)

			return responseForm.User, responseForm.Password
		}
		resp.Body.Close()
		time.Sleep(5 * time.Second)
	}
}

func Bootstrap(uuid string, tenant string, password string) (string, error) {
	// configure MQTT client
	address := "tcps://" + tenant + ".cumulocity.com:8883/" // scheme://host:port
	log.Println(address)
	opts := MQTT.NewClientOptions().AddBroker(address)
	opts.SetUsername("management/devicebootstrap")
	opts.SetPassword(password)
	opts.SetClientID(uuid)

	c8yError := make(chan error)
	c8yAuth := make(chan string)

	answerReceived := false

	// Answer receive-callback
	authReceive := func(client MQTT.Client, msg MQTT.Message) {
		answerReceived = true
		answer := string(msg.Payload())
		if strings.HasPrefix(answer, "70") {
			log.Println("received authorization: " + answer)
			c8yAuth <- answer
		} else {
			log.Println("received unknown message:" + answer)
			c8yError <- errors.New(fmt.Sprintf("unknown message received: %v", msg))
		}
	}

	// Error receive-callback
	errReceive := func(client MQTT.Client, msg MQTT.Message) {
		answerReceived = true
		answer := string(msg.Payload())
		log.Println("received error message:" + answer)
		c8yError <- errors.New(fmt.Sprintf("error message received: %v", msg))
	}

	// configure OnConnect callback: subscribe
	opts.OnConnect = func(c MQTT.Client) {
		log.Println("MQTT client connected.")

		// subscribe to authorization message
		if token := c.Subscribe("s/dcr", 0, authReceive); token.Wait() && token.Error() != nil {
			c8yError <- token.Error()
			return
		}

		// subscribe to error messages
		if token := c.Subscribe("s/e", 0, errReceive); token.Wait() && token.Error() != nil {
			c8yError <- token.Error()
			return
		}

		// publish until answer received
		for !answerReceived {
			log.Println("publishing...")
			c.Publish("s/ucr", 0, false, nil)
			time.Sleep(5 * time.Second)
		}
	}

	client := MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return "", token.Error()
	}

	defer client.Disconnect(0)

	select {
	case auth := <-c8yAuth:
		return auth, nil
	case err := <-c8yError:
		return "", err
	}
}

func GetClient(uuid string, tenant string, user string, password string) (MQTT.Client, error) {
	address := "tcps://" + tenant + ".cumulocity.com:8883/"
	opts := MQTT.NewClientOptions().AddBroker(address)
	opts.SetClientID(uuid)
	opts.SetUsername(tenant + "/" + user)
	opts.SetPassword(password)

	c8yError := make(chan error)
	c8yReady := make(chan bool)

	// callback for error messages
	receive := func(client MQTT.Client, msg MQTT.Message) {
		answer := string(msg.Payload())
		log.Println("received error message:" + answer)
	}

	// configure OnConnect callback: subscribe to error messages when connected
	opts.OnConnect = func(c MQTT.Client) {
		log.Println("MQTT client connected.")

		// subscribe to error messages
		if token := c.Subscribe("s/e", 0, receive); token.Wait() && token.Error() != nil {
			c8yError <- token.Error()
		} else {
			c8yReady <- true
		}
	}

	// create client and connect
	client := MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}

	// wait for OnConnect callback
	select {
	case _ = <-c8yReady:
		return client, nil
	case err := <-c8yError:
		return nil, err
	}
}

func Send(c MQTT.Client, valueToSend bool) error {
	var message string
	if valueToSend {
		// send true (1)
		message = "200,c8y_Bool,B,1"
	} else {
		// send false (0)
		message = "200,c8y_Bool,B,0"
	}

	if token := c.Publish("s/us", 0, false, message); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}
