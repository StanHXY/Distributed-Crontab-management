package master

import (
	"net"
	"net/http"
	"time"
)

// http API for tasks
type ApiServer struct {
	httpServer *http.Server
}

var (
	G_apiServer *ApiServer
)

// save tasks API
func handleJobSave(w http.ResponseWriter, r *http.Request) {

}

// initialize service
func InitApiServer() (err error) {
	var (
		mux        *http.ServeMux
		listener   net.Listener
		httpServer *http.Server
	)
	// config router
	mux = http.NewServeMux()
	mux.HandleFunc("/job/save", handleJobSave)

	// launch TCP Listener
	if listener, err = net.Listen("TCP", ":8070"); err != nil {
		return
	}

	// Create Http service
	httpServer = &http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		Handler:      mux,
	}

	G_apiServer = &ApiServer{
		httpServer: httpServer,
	}

	// launch the Service
	go httpServer.Serve(listener)

	return

}
