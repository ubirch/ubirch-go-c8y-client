package c8y

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/eclipse/paho.mqtt.golang"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

func bootstrapHTTP(uuid string, tenant string, password string) (map[string]string, error) {
	log.Println("Bootstrapping...")
	data, err := json.Marshal(map[string]string{"id": uuid})
	if err != nil {
		return nil, err
	}
	url := "https://" + tenant + ".cumulocity.com/devicecontrol/deviceCredentials"
	client := http.Client{}

	for {
		request, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
		if err != nil {
			return nil, err
		}
		request.Header.Set("Content-Type", "application/vnd.com.nsn.cumulocity.deviceCredentials+json")
		request.Header.Set("Accept", "application/vnd.com.nsn.cumulocity.deviceCredentials+json")
		request.SetBasicAuth("devicebootstrap", password)

		resp, err := client.Do(request)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode == http.StatusNotFound {
			log.Print("Acceptance pending: " + uuid)
		} else if resp.StatusCode == http.StatusCreated {
			deviceCredentials := make(map[string]string)

			// read response body
			bodyBytes, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}
			bodyString := string(bodyBytes)
			log.Println(bodyString)

			// parse json response body
			err = json.Unmarshal(bodyBytes, &deviceCredentials)
			if err != nil {
				return nil, err
			}
			resp.Body.Close()

			return deviceCredentials, nil
		} else {
			log.Println("Response status:", resp.Status)
		}
		resp.Body.Close()
		time.Sleep(5 * time.Second)
	}
}

func getCredentials(uuid string, tenant string, bootstrapPW string) (map[string]string, error) {
	var deviceCredentials map[string]string
	// check for device credentials file
	credentialsFilename := uuid + ".json"
	_, err := os.Stat(credentialsFilename)
	if os.IsNotExist(err) { // file does not exist
		// bootstrap
		deviceCredentials, err = bootstrapHTTP(uuid, tenant, bootstrapPW)
		if err != nil {
			return nil, err
		}
		// create file and save credentials to it
		deviceCredentialsJson, err := json.Marshal(deviceCredentials)
		if err != nil {
			return nil, err
		}
		credentialsFile, err := os.Create(credentialsFilename)
		if err != nil {
			return nil, err
		}
		defer credentialsFile.Close()

		_, err = credentialsFile.Write(deviceCredentialsJson)
		if err != nil {
			return nil, err
		}
		log.Printf("created Cumulocity device credentials file: %s \n", credentialsFilename)
	} else { // file exists
		log.Println("reading credentials from file")
		deviceCredentialsJson, err := ioutil.ReadFile(credentialsFilename)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(deviceCredentialsJson, &deviceCredentials)
		if err != nil {
			return nil, err
		}
	}
	return deviceCredentials, nil
}

func GetClient(uuid string, tenant string, bootstrapPW string) (mqtt.Client, error) {
	deviceCredentials, err := getCredentials(uuid, tenant, bootstrapPW)
	if err != nil {
		return nil, err
	}
	// initialize MQTT client
	address := "tcps://" + tenant + ".cumulocity.com:8883/"
	opts := mqtt.NewClientOptions().AddBroker(address)
	opts.SetClientID(uuid)
	opts.SetUsername(tenant + "/" + deviceCredentials["username"])
	opts.SetPassword(deviceCredentials["password"])

	c8yError := make(chan error)
	c8yReady := make(chan bool)

	// callback for error messages
	receive := func(client mqtt.Client, msg mqtt.Message) {
		c8yError := string(msg.Payload())
		if c8yError == "41,100,Device already existing" || strings.HasPrefix(c8yError, "50,100,") {
			return
		}
		log.Println("MQTT client received error message from Cumulocity: " + c8yError)
	}

	// configure OnConnect callback: subscribe to error messages when connected
	opts.OnConnect = func(c mqtt.Client) {
		log.Println("MQTT client connected.")

		// subscribe to error messages
		if token := c.Subscribe("s/e", 0, receive); token.Wait() && token.Error() != nil {
			c8yError <- token.Error()
			return
		}

		// create device identity in cumulocity registry
		deviceCreationMsg := "100," + uuid + ",c8y_MQTTDevice" // "100,Device Name,Device Type"
		if token := c.Publish("s/us", 0, false, deviceCreationMsg); token.Wait() && token.Error() != nil {
			c8yError <- token.Error()
		} else {
			c8yReady <- true
		}
	}

	// create client and connect
	client := mqtt.NewClient(opts)
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

func Send(c mqtt.Client, name string, value byte, timestamp time.Time) error {
	const timeFormat = "2006-01-02T15:04:05.000Z"

	message := fmt.Sprintf("200,c8y_Switch,%s,%d,,%s", name, value, timestamp.Format(timeFormat))
	log.Println("sending message to Cumulocity: " + message)

	if token := c.Publish("s/us", 0, false, message); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}
