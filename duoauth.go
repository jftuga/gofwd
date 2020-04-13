package main

import (
	"fmt"
	"net/url"
	"os"

	duoapi "github.com/duosecurity/duo_api_golang"
	"github.com/duosecurity/duo_api_golang/authapi"
	"gopkg.in/ini.v1"
)

func duoReadConfig(cfgFile string) {
	cfg, err := ini.Load(cfgFile)
	if err != nil {
		fmt.Printf("Fail to read file: %v", err)
		os.Exit(1)
	}
	fmt.Println(cfg.SectionStrings())
}

func duoCheck() bool {
	duoClient := duoapi.NewDuoApi("integrationKey", "secretKey", "apiHostname", "go-client")
	if duoClient == nil {
		fmt.Println("Error #100: Failed to create new Duo Api")
		return false
	}
	duoAuthClient := authapi.NewAuthApi(*duoClient)
	check, err := duoAuthClient.Check()
	if err != nil {
		fmt.Println("Error #150:", err)
		return false
	}
	if check == nil {
		fmt.Println("Error #155: 'check' is nil")
		return false
	}

	var msg, detail string
	if check.StatResult.Message != nil {
		msg = *check.StatResult.Message
	}
	if check.StatResult.Message_Detail != nil {
		detail = *check.StatResult.Message_Detail
	}
	if check.StatResult.Stat != "OK" {
		fmt.Printf("Error #180: Could not connect to Duo: %q (%q)", msg, detail)
		return false
	}

	duoUser := "duoUser"
	options := []func(*url.Values){authapi.AuthUsername(duoUser)}
	options = append(options, authapi.AuthDevice("auto"))
	result, err := duoAuthClient.Auth("push", options...)
	if err != nil {
		fmt.Println("Error #200:", err)
		return false
	}
	if result == nil {
		fmt.Println("Error #220: 'result' is nil")
		return false
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

	fmt.Println("final verdict:", success)
	if !success {
		return false
	}

	return true
}
