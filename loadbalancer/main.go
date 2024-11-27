package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type Server interface {
	Address() string
	ISAlive() bool
	Serve(w http.ResponseWriter, r *http.Request)
}

type simpleServer struct {
	address string
	proxy   *httputil.ReverseProxy
}

type loadbalancer struct {
	port            string
	roundRobinCount int
	servers         []Server
}

func handleError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func newSimpleServer(address string) *simpleServer {
	serverUrl, err := url.Parse(address)
	handleError(err)
	proxy := httputil.NewSingleHostReverseProxy(serverUrl)
	return &simpleServer{
		address: address,
		proxy:   proxy,
	}
}

func newLoadbalancer(port string, servers []Server) *loadbalancer {
	return &loadbalancer{
		port:            port,
		roundRobinCount: 0,
		servers:         servers,
	}
}

func (s *simpleServer) Address() string {
	return s.address
}

func (s *simpleServer) ISAlive() bool {
	return true
}

func (s *simpleServer) Serve(w http.ResponseWriter, r *http.Request) {
	s.proxy.ServeHTTP(w, r)
}

func (lb *loadbalancer) getNextAvailableServer() Server {
	server := lb.servers[lb.roundRobinCount%len(lb.servers)]
	if !server.ISAlive() {
		lb.roundRobinCount++
		server = lb.servers[lb.roundRobinCount%len(lb.servers)]
	}
	lb.roundRobinCount++
	return server
}

func (lb *loadbalancer) ServeProxy(w http.ResponseWriter, r *http.Request) {
	nextServer := lb.getNextAvailableServer()
	nextServer.Serve(w, r)
}

func main() {
	// Create servers
	servers := []Server{
		newSimpleServer("http://gmail.com"),
		newSimpleServer("http://google.com"),
		newSimpleServer("http://notion.com"),
	}

	// Create load balancer
	lb := newLoadbalancer("5000", servers)

	handleRedirect := func(w http.ResponseWriter, r *http.Request) {
		lb.ServeProxy(w, r)
	}

	http.HandleFunc("/", handleRedirect)
	fmt.Printf("Listening on port %s\n", lb.port)
	err := http.ListenAndServe(":"+lb.port, nil)
	handleError(err)
}
