package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"strings"
)

type wwwServiceConfiguration struct {
	Ip	string					`json:"ip"`
	SecurePortNumber int		`json:"secureportnumber"`	// (ie: 443)
	InsecurePortNumber int 		`json:"insecureportnumber"`	// These will ALWAYS be redirected to SecurePortNumber (ie: 80 redirected to 443)
}

func startWWWService(channel chan string) {
	webConfig,err := loadConfig()
	if err != nil { channel <- "Could Not Start WWW Server: " + err.Error(); return}
	logger.Printf("*** Starting WWW Service on %s:%d [Redirect From %d]\r\n",webConfig.Ip,webConfig.SecurePortNumber,webConfig.InsecurePortNumber)
	// Do stuff, catch quit
	for true {
		select {
			case v := <-channel:
				logger.Printf("[WWW Service] Signal Received: %s",v)
				if strings.EqualFold(v,"Shutdown") {
					shutdownWWWService(channel,webConfig)
				}
			default:
		}
	}
}

func shutdownWWWService(channel chan string,webConfig wwwServiceConfiguration) {
	logger.Printf("*** Shutting Down WWW Service on %s:%d [Redirect From %d]\r\n",webConfig.Ip,webConfig.SecurePortNumber,webConfig.InsecurePortNumber)
	channel <- "Fine."
}

// Config
func loadConfig() (wwwServiceConfiguration,error) {
	// Check current directory for ./goCryptoMarketSite.json
	checkFile,err := fileExists("goCryptoMarketSite.json")
	if err != nil { return wwwServiceConfiguration{},err }
	// Check /etc/goCryptoMarketSite.json
	if checkFile {
		// Load and Return
		logger.Println("Loading ./goCryptoMarketSite.json")
		file,_ := ioutil.ReadFile("goCryptoMarketSite.json")
        conf := wwwServiceConfiguration{}
        err := json.Unmarshal([]byte(string(file)), &conf)
        if err != nil { return wwwServiceConfiguration{},err }
        return conf,nil
	} else {
		checkEtc, err := fileExists("/etc/goCryptoMarketSite.json")
		if err != nil { return wwwServiceConfiguration{}, err }
		if checkEtc {
			// Load and Return
			logger.Println("Loading /etc/goCryptoMarketSite.json")
			file,_ := ioutil.ReadFile("/etc/goCryptoMarketSite.json")
			conf := wwwServiceConfiguration{}
			json.Unmarshal([]byte(string(file)), &conf)
			if err != nil { return wwwServiceConfiguration{},err }
			return conf,nil
		}
	}
	// Give up and quit
	return wwwServiceConfiguration{},errors.New("Could not find goCryptoMarketSite.json (local or /etc)")
}

// Misc
func fileExists(path string) (bool, error) {
    _, err := os.Stat(path)
    if err == nil { return true, nil }
    if os.IsNotExist(err) { return false, nil }
    return true, err
}