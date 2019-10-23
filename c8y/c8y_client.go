package c8y

import (
	"bytes"
	"encoding/json"
	"github.com/eclipse/paho.mqtt.golang"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

func BootstrapHTTP(uuid string, tenant string, password string) map[string]string {
	log.Println("Bootstrapping")
	data, err := json.Marshal(map[string]string{"id": uuid})
	if err != nil {
		panic(err)
	}
	url := "https://" + tenant + ".cumulocity.com/devicecontrol/deviceCredentials"
	client := http.Client{}

	for {
		request, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
		if err != nil {
			panic(err)
		}
		request.Header.Set("Content-Type", "application/vnd.com.nsn.cumulocity.deviceCredentials+json")
		request.Header.Set("Accept", "application/vnd.com.nsn.cumulocity.deviceCredentials+json")
		request.SetBasicAuth("devicebootstrap", password)

		resp, err := client.Do(request)
		if err != nil {
			log.Fatalf(err.Error())
		}

		log.Println("Response status:", resp.Status)

		if resp.StatusCode == http.StatusCreated {
			deviceCredentials := make(map[string]string)

			// read response body
			bodyBytes, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Fatalf("ERROR: unable to read response: %v", err)
			}
			bodyString := string(bodyBytes)
			log.Println(bodyString)

			// parse json response body
			err = json.Unmarshal(bodyBytes, &deviceCredentials)
			if err != nil {
				log.Fatalf("ERROR: unable to parse response: %v", err)
			}
			resp.Body.Close()

			// get username and password from response
			return deviceCredentials
		}
		resp.Body.Close()
		time.Sleep(10 * time.Second)
	}
}

func GetClient(uuid string, tenant string, bootstrapPW string) (mqtt.Client, error) {
	var deviceCredentials map[string]string
	// check for device credentials file
	credentialsFilename := uuid + ".ini"
	_, err := os.Stat(credentialsFilename)
	if os.IsNotExist(err) { // file does not exist
		// bootstrap
		deviceCredentials = BootstrapHTTP(uuid, tenant, bootstrapPW)
		deviceCredentialsJson, err := json.Marshal(deviceCredentials)
		if err != nil {
			panic(err)
		}

		// create JSON file and write credentials to it
		jsonFile, err := os.Create(credentialsFilename)
		if err != nil {
			panic(err)
		}
		defer jsonFile.Close()

		n, err := jsonFile.Write(deviceCredentialsJson)
		log.Printf("wrote %d bytes to %s \n", n, credentialsFilename)

	} else { // file exists
		log.Println("reading credentials from file")
		deviceCredentialsJson, err := ioutil.ReadFile(credentialsFilename)
		if err != nil {
			log.Fatalf("ERROR: unable to read device credentials file: %v", err)
		}
		err = json.Unmarshal(deviceCredentialsJson, &deviceCredentials)
		if err != nil {
			log.Fatalf("ERROR: unable to parse device credentials: %v", err)
		}
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
		answer := string(msg.Payload())
		log.Println("received error message: " + answer)
	}

	// configure OnConnect callback: subscribe to error messages when connected
	opts.OnConnect = func(c mqtt.Client) {
		log.Println("MQTT client connected.")

		// subscribe to error messages
		if token := c.Subscribe("s/e", 0, receive); token.Wait() && token.Error() != nil {
			c8yError <- token.Error()
		} else {
			c8yReady <- true
		}

		// create device identity in cumulocity registry
		deviceCreationMsg := "100," + uuid + ",c8y_MQTTDevice" // "100,Device Name,Device Type"
		if token := c.Publish("s/us", 0, false, deviceCreationMsg); token.Wait() && token.Error() != nil {
			c8yError <- token.Error()
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

func Send(c mqtt.Client, valueToSend bool) error {
	log.Println("sending...")
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
