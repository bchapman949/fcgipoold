package fcgipoold

import (
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

// Pool ..
type Pool struct {
	queue     chan Work
	startport int
	numworker int
	command   string
	reuse     int
}

// Work is sent through a channel
// to the client-process
type Work struct {
	Req     *http.Request
	Result  chan *http.Response
	Retries int
}

// NewPool creates a new worker pool. The child-processes
// are spawned and maintained automatically.
func NewPool(cmd string, startp, numw, reuse int) *Pool {
	p := &Pool{
		queue:     make(chan Work),
		startport: startp,
		reuse:     reuse,
		numworker: numw,
		command:   cmd,
	}
	p.run()
	return p
}

// run ..
func (p *Pool) run() {
	for i := p.startport; i < p.startport+p.numworker; i++ {
		go func(i int) {
			worker := newWorker(p.command, strconv.Itoa(i), p.reuse)
			for {
				select {
				case msg := <-p.queue:
					worker.work(msg)
				}
			}
		}(i)
	}
}

// Dispatch..
func (p *Pool) Dispatch(res http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	back := make(chan *http.Response)
	p.queue <- Work{
		Req:    req,
		Result: back,
	}
	fcgiRes := <-back
	if fcgiRes.Body != nil {
		content, err := ioutil.ReadAll(fcgiRes.Body)
		if err != nil {
			log.Printf("error reading proxy response body: %v", err)
		}
		res.WriteHeader(200)
		res.Write(content)
		return
	}
	res.WriteHeader(500)
}
