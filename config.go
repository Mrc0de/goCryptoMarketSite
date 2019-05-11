package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
)

type wwwServiceConfiguration struct {
	Ip	string					`json:"ip"`
	SecurePortNumber int		`json:"secureportnumber"`	// (ie: 443)
	InsecurePortNumber int 		`json:"insecureportnumber"`	// This will ALWAYS be redirected to SecurePortNumber
	CertFile string				`json:"certfile"`
	KeyFile	string				`json:"keyfile"`
	WWWServiceHostname string	`json:"wwwservicehostname"`	// Required, Ignore all requests not for this hostname (or www.hostname )
															// This will eventually become a list to support multiple hostnames
}

// Config
func loadConfig() (wwwServiceConfiguration,error) {
	// There is a bug in how I do this,
	// Total Fail if the file isn't found, for now, just make sure it exists (and is filled in)
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
	logger.Printf("Never")
	return wwwServiceConfiguration{}, errors.New("This Should Never Happen.")
}