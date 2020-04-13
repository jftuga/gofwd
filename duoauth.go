package main

import (
	"fmt"
	"net/url"

	duoapi "github.com/duosecurity/duo_api_golang"
	"github.com/duosecurity/duo_api_golang/authapi"
	"gopkg.in/ini.v1"
)

type duoCredentials struct {
	name        string
	integration string
	secret      string
	hostname    string
}

func duoReadConfig(cfgFile string, name string) (duoCredentials, error) {
	var duoCred duoCredentials
	var err error
	cfg, err := ini.Load(cfgFile)
	if err != nil {
		err = fmt.Errorf("Fail to read file: %v", err)
		return duoCred, err
	}

	sectionType := cfg.Section(name).Key("type").String()
	if "duo" == sectionType {
		duoCred.name = name
		duoCred.integration = cfg.Section(name).Key("integration").String()
		duoCred.secret = cfg.Section(name).Key("secret").String()
		duoCred.hostname = cfg.Section(name).Key("hostname").String()
	}

	if 0 == len(duoCred.name) {
		err = fmt.Errorf("[%s] Duo Config: Invalid user name", name)
	} else if len(duoCred.integration) < 15 {
		err = fmt.Errorf("[%s] Duo Config: Invalid integration", name)
	} else if len(duoCred.secret) < 15 {
		err = fmt.Errorf("[%s] Duo Config: Invalid secret", name)
	} else if len(duoCred.hostname) < 15 {
		err = fmt.Errorf("[%s] Duo Config: Invalid hostname", name)
	}

	return duoCred, err
}

func duoCheck(duoCred duoCredentials) (bool, error) {
	var err error

	duoClient := duoapi.NewDuoApi(duoCred.integration, duoCred.secret, duoCred.hostname, "go-client")
	if duoClient == nil {
		err = fmt.Errorf("Error #100: Failed to create new Duo Api")
		return false, err
	}
	duoAuthClient := authapi.NewAuthApi(*duoClient)
	check, err := duoAuthClient.Check()
	if err != nil {
		err = fmt.Errorf("Error #150: %s", err)
		return false, err
	}
	if check == nil {
		err = fmt.Errorf("Error #155: 'check' is nil")
		return false, err
	}

	var msg, detail string
	if check.StatResult.Message != nil {
		msg = *check.StatResult.Message
	}
	if check.StatResult.Message_Detail != nil {
		detail = *check.StatResult.Message_Detail
	}
	if check.StatResult.Stat != "OK" {
		err = fmt.Errorf("Error #180: Could not connect to Duo: %q (%q)", msg, detail)
		return false, err
	}

	duoUser := duoCred.name
	options := []func(*url.Values){authapi.AuthUsername(duoUser)}
	options = append(options, authapi.AuthDevice("auto"))
	result, err := duoAuthClient.Auth("push", options...)
	if err != nil {
		err = fmt.Errorf("Error #200: %s", err)
		return false, err
	}
	if result == nil {
		err = fmt.Errorf("Error #220: 'result' is nil")
		return false, err
	}

	if false {
		fmt.Println("result:", result)
		fmt.Println("-----------------------")
		fmt.Println("result.StatResult:", result.StatResult)
		fmt.Println("-----------------------")
		fmt.Println("result.Response:", result.Response)
	}

	success := false
	if result.StatResult.Stat == "OK" && result.Response.Result == "allow" {
		success = true
	}

	if !success {
		err = fmt.Errorf("Error #230: 'success' is false")
		return false, err
	}

	return true, nil
}
