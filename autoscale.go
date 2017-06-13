package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	cfclient "github.com/cloudfoundry-community/go-cfclient"
	"github.com/prometheus/common/log"
)

func scale(appName string, instances int) {
	cfPassword := os.Getenv("PASSWORD")
	c := &cfclient.Config{
		ApiAddress:        "http://api.sys.bogata.cf-app.com",
		Username:          "admin",
		Password:          cfPassword,
		SkipSslValidation: true,
	}

	client, err := cfclient.NewClient(c)
	if err != nil {
		fmt.Println(err)
	}

	spaceGuid := "27d5badd-3ba9-43c8-a96a-ca572a801140"
	orgGuid := "0e88e8d8-dc52-4f71-ba69-ca1388000f53"

	app, err := client.AppByName(appName, spaceGuid, orgGuid)
	if err != nil {
		log.Infof("app not found %s", appName)
	}

	scale := Scale{instances}

	b := new(bytes.Buffer)
	err = json.NewEncoder(b).Encode(scale)
	if err != nil {
		log.Errorf("error: %v", err)
	}

	req := client.NewRequestWithBody("PUT", fmt.Sprintf("/v2/apps/%s", app.Guid), b)

	_, err = client.DoRequest(req)
	if err != nil {
		log.Errorf("error making request %+v\n", err)
	}

	log.Infof("Scaled app '%s' to %d instances", appName, instances)
}

type Scale struct {
	Instances int `json:"instances"`
}
