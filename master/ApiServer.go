package master

import (
	"net"
	"net/http"
	"strconv"
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
	if listener, err = net.Listen("tcp", ":"+strconv.Itoa(G_config.ApiPort)); err != nil {
		return
	}

	// Create Http service
	httpServer = &http.Server{
		ReadTimeout:  time.Duration(G_config.ApiReadTimeout) * time.Millisecond,
		WriteTimeout: time.Duration(G_config.ApiWriteTimeout) * time.Millisecond,
		Handler:      mux,
	}

	G_apiServer = &ApiServer{
		httpServer: httpServer,
	}

	// launch the Service
	go httpServer.Serve(listener)

	return

}
