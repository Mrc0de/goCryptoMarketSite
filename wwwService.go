package main

import (
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
	"strings"
)


//////////
// Startup
func startWWWService(channel chan string,webConfig wwwServiceConfiguration) {
	logger.Printf("*** Starting WWW Service on %s:%d [Redirect From %d]\r\n",webConfig.Ip,
								webConfig.SecurePortNumber,webConfig.InsecurePortNumber)
	// Do stuff, catch quit
	r := mux.NewRouter()
	r.HandleFunc("/",wwwHome)

	/////////////
	go startSecure(webConfig,r)
	go startInsecure(webConfig)
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


////////////////
// StartInsecure
func startInsecure(webConfig wwwServiceConfiguration){
	// This STARTS the redirect to securePort.
	// If this fails, we will exit the application (Panic)
	logger.Printf("Starting Insecure Port Redirect To Secure Port Listener...")
	err := http.ListenAndServe(webConfig.Ip + ":" + strconv.Itoa(webConfig.InsecurePortNumber),
		http.HandlerFunc(func(w http.ResponseWriter,req *http.Request){
			logger.Printf("[%s] Redirecting %s from %s to %s",req.RequestURI,req.RemoteAddr,
									strconv.Itoa(webConfig.InsecurePortNumber),strconv.Itoa(webConfig.SecurePortNumber))
			http.Redirect(w,req,"https://"+req.Host+req.URL.String(),http.StatusSeeOther) // 303 doesnt get cached (usually)
	}))
	if err != nil {
		logger.Printf("Error Starting Insecure Port Redirect to Secure Port Listener: %s",err);
		logger.Panic("Quitting.")
	}
}


//////////////
// StartSecure
func startSecure(webConfig wwwServiceConfiguration,r *mux.Router){
	// This STARTS the actual secure server (using the configured cert/key combo)
	// If this fails, we will exit the application (Panic)
	logger.Printf("Starting Secure Port Listener...")
	err := http.ListenAndServeTLS(webConfig.Ip+ ":" + strconv.Itoa(webConfig.SecurePortNumber),webConfig.CertFile,
		webConfig.KeyFile,r)
	if err != nil {
		logger.Printf("Error Starting Secure Port Listener: %s",err);
		logger.Panic("Quitting.")
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
	intRet,err := w.Write([]byte("w00t! - Secure"))
	if err != nil {
		logger.Printf("Error Writing reply to [%s]: [%d] %s",r.RemoteAddr,intRet,err)
	}
}