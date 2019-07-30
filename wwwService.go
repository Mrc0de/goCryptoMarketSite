package main

import (
	"github.com/gorilla/mux"
	"html/template"
	"net/http"
	"strconv"
	"strings"
)


//////////
// Startup
func startWWWService(channel chan string,webConfig wwwServiceConfiguration) {
	logger.Printf("*** Starting WWW Service on %s:%d [Redirect From %d]\r\n",webConfig.Ip,
								webConfig.SecurePortNumber,webConfig.InsecurePortNumber)
	hub := newWsHub()
	go hub.run()
	r := mux.NewRouter()
	r.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		websocketUpgrade(hub, w, r)
	})
	logger.Println("*** Starting WebSocket Service on WWWService at /ws")
	r.PathPrefix("/imgs/").Handler(http.StripPrefix("/imgs/", http.FileServer(http.Dir("./imgs/"))))
	r.HandleFunc("/",wwwHome)

	// Image Content
	////////////////
	go startSecure(webConfig,r)
	go startInsecure(webConfig)
	///////////////////////////

	for  {
		select {
			case v := <-channel: {
				logger.Printf("[WWW Service] Signal Received: %s", v)
				if strings.EqualFold(v, "Shutdown") {
					shutdownWWWService(channel, webConfig)
				}
			}
			default: {
				//
			}
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
			http.Redirect(w,req,"https://"+req.Host+req.URL.String(),http.StatusFound) // 302 doesnt get cached (usually)
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
	var data struct {
		WsHost string
	}
	data.WsHost = r.Host
	logger.Printf("[%s] %s %s",r.RequestURI,r.Method,r.RemoteAddr)
	tmpl := template.Must(template.ParseFiles("templates/Home.tmpl","templates/Base.tmpl"))
	err := tmpl.Execute(w,&data)
	if err != nil { logger.Printf("Error Parsing Template: %s",err) }
}