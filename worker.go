package fcgipoold

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/fcgi"
	"os/exec"
	"strconv"
	"time"

	"github.com/tomasen/fcgi_client"
)

var httpInternalError = &http.Response{StatusCode: http.StatusInternalServerError}

func newWorker(command, port string, max int) *fcgiChild {
	worker := &fcgiChild{
		command: command,
		port:    port,
		counter: 0,
		exitc:   make(chan struct{}),
		max:     max,
	}
	worker.spawn()
	return worker
}

type fcgiChild struct {
	port    string
	command string
	cmd     *exec.Cmd
	outpipe io.ReadCloser
	exitc   chan struct{}
	counter int
	max     int
}

func (c *fcgiChild) work(msg Work) {
	defer c.respawn()
	c.counter++
	env := fcgi.ProcessEnv(msg.Req)
	env["SERVER_PROTOCOL"] = msg.Req.Proto
	env["REQUEST_METHOD"] = msg.Req.Method
	client, err := fcgiclient.Dial("tcp", "127.0.0.1:"+c.port)
	if err != nil {
		log.Printf("err %v", err)
		<-time.After(time.Millisecond * 10)
		if msg.Retries >= 10 {
			msg.Result <- httpInternalError
			return
		}
		msg.Retries++
		c.work(msg)
		return
	}
	if msg.Req.Method == http.MethodGet {
		res, err := client.Get(env)
		if err != nil {
			log.Printf("err: %v", err)
			msg.Retries++
			c.work(msg)
			return
		}
		msg.Result <- res
		return
	}

	if msg.Req.Method == http.MethodPost {
		body, err := ioutil.ReadAll(msg.Req.Body)
		if err != nil {
			log.Printf("err: %v", err)
			msg.Result <- httpInternalError
			return
		}
		rd := bytes.NewBuffer(body)
		len, err := strconv.Atoi(msg.Req.Header.Get("Content-Length"))
		if err != nil {
			len = 0
		}
		res, err := client.Post(env, msg.Req.Header.Get("Content-Type"), rd, len)
		if err != nil {
			log.Printf("err: %v", err)
			msg.Result <- httpInternalError
			return
		}
		msg.Result <- res
		return
	}
}

func (c *fcgiChild) spawn() {
	c.cmd = exec.Command(c.command, c.port)
	c.outpipe, _ = c.cmd.StderrPipe()
	err := c.cmd.Start()
	if err != nil {
		log.Printf("err: %v", err)
	}
	c.wait()
}

func (c *fcgiChild) respawn() {
	if c.counter >= c.max {
		c.counter = 0
		err := c.cmd.Process.Kill()
		if err != nil {
			log.Printf("err: %v", err)
		}
		c.spawn()
	}
}

// this func waits for a newline on stderr
// that way the process tells us
// that he is ready
func (c *fcgiChild) wait() {
	if c.cmd == nil {
		return
	}
	defer c.outpipe.Close()
	rd := bufio.NewReader(c.outpipe)
	rd.ReadString('\n')
}
