package main

import (
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
	"strings"
)


//////////
// Startup
func startWWWService(channel chan string) {
	webConfig,err := loadConfig()
	if err != nil { channel <- "Could Not Start WWW Server: " + err.Error(); return}
	logger.Printf("*** Starting WWW Service on %s:%d [Redirect From %d]\r\n",webConfig.Ip,
								webConfig.SecurePortNumber,webConfig.InsecurePortNumber)
	// Do stuff, catch quit
	r := mux.NewRouter()
	r.HandleFunc("/",wwwHome)

	/////////////
	go http.ListenAndServe(webConfig.Ip + ":" + strconv.Itoa(webConfig.InsecurePortNumber),
		http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
			http.Redirect(w,r,"https://"+r.Host+r.URL.String(),http.StatusMovedPermanently)
	}))
	/////////////
	go http.ListenAndServeTLS(webConfig.Ip+ ":" + strconv.Itoa(webConfig.SecurePortNumber),webConfig.CertFile,
		webConfig.KeyFile,r)
	/////////////
	for  {
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


///////////
// Shutdown
func shutdownWWWService(channel chan string,webConfig wwwServiceConfiguration) {
	logger.Printf("*** Shutting Down WWW Service on %s:%d [Redirect From %d]\r\n",webConfig.Ip,
							webConfig.SecurePortNumber,webConfig.InsecurePortNumber)
	channel <- "Fine."
}

///////////
// Home "/"
func wwwHome(w http.ResponseWriter, r *http.Request) {
	logger.Printf("[%s] %s %s",r.RequestURI,r.Method,r.RemoteAddr)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("w00t! - Secure"))
}