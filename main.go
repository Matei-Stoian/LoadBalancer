package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"sync"
	"sync/atomic"
)

type LoadBalancer struct {
	targets         []*url.URL
	currentIndex    int
	maxGRoutine     int
	currentGRoutine int32
	mux             sync.Mutex
	wg              sync.WaitGroup
	requestQueue    chan *http.Request
}

func NewLoadBalancer(target []*url.URL, maxGRoutines int, queuSize int) *LoadBalancer {
	return &LoadBalancer{
		targets:         target,
		maxGRoutine:     maxGRoutines,
		currentGRoutine: 0,
		requestQueue:    make(chan *http.Request, queuSize),
	}
}
func parseDocker(env string) []*url.URL {
	backendEnv := os.Getenv(env)
	backendServers := strings.Split(backendEnv, ",")
	if len(backendServers) == 0 {
		log.Fatal("No backend servers listed")
	}
	var parsedTargets []*url.URL
	for _, server := range backendServers {
		parseServer, err := url.Parse(server)
		if err != nil {
			log.Fatal("Failed to parse: ", server)
		}
		parsedTargets = append(parsedTargets, parseServer)
	}
	return parsedTargets
}
func (lb *LoadBalancer) getNext() *url.URL {
	lb.mux.Lock()
	defer lb.mux.Unlock()
	target := lb.targets[lb.currentIndex]
	lb.currentIndex = (lb.currentIndex + 1) % len(lb.targets)
	return target
}
func (lb *LoadBalancer) healtCheck(target *url.URL) bool {
	resp, err := http.Get(target.String())
	if err != nil || resp.StatusCode != http.StatusOK {
		return false
	}
	return true
}
func (lb *LoadBalancer) ProccesRequest(target *url.URL, w http.ResponseWriter, r *http.Request) {
	atomic.AddInt32(&lb.currentGRoutine, 1)
	defer atomic.AddInt32(&lb.currentGRoutine, -1)

	healthOk := lb.healtCheck(target)
	if !healthOk {

		return
	}
	proxy := NewReverseProxy(target)
	proxy.ServeHTTP(w, r)
}
func (lb *LoadBalancer) handle(w http.ResponseWriter, r *http.Request) {
	if atomic.LoadInt32(&lb.currentGRoutine) >= int32(lb.maxGRoutine) {
		lb.requestQueue <- r
		http.Error(w, "Request queued", http.StatusProcessing)
		return
	}
	lb.wg.Add(1)
	defer lb.wg.Done()
	target := lb.getNext()
	go lb.ProccesRequest(target, w, r)
}
func NewReverseProxy(target *url.URL) *httputil.ReverseProxy {
	director := func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path = target.Path
		req.Host = target.Host
	}
	return &httputil.ReverseProxy{Director: director}
}
func (lb *LoadBalancer) proccesQueue() {
	for {
		select {
		case req := <-lb.requestQueue:
			if req != nil {
				target := lb.getNext()

				go lb.ProccesRequest(target, nil, req)
			}
		}

	}
}
func main() {
	backendServers := parseDocker("BACKEND_SERVERS")
	maxGRoutines := 60
	queueSize := 100

	lb := NewLoadBalancer(backendServers, maxGRoutines, queueSize)
	go lb.proccesQueue()
	http.HandleFunc("/", lb.handle)
	port := 8080
	for _, v := range lb.targets {
		fmt.Println(v.Host)
	}
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		log.Fatal("Error starting the server: ", err)
	}
	lb.wg.Wait()
}
