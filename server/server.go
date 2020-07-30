package server

import (
	"bufio"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
)

// Server anyconnect server
type Server struct {
	Listeners []net.Listener
	PublicDir string
	Loger     *log.Logger

	htmlFS http.FileSystem

	run       bool
	mux       *http.ServeMux
	waitGroup *sync.WaitGroup
	closeChan chan struct{}
}

// Serve 开始工作
func (s *Server) Serve() error {

	if s.Listeners == nil || len(s.Listeners) == 0 {
		return errors.New("no listener set")
	}

	s.htmlFS = http.Dir(s.PublicDir)
	s.mux = http.NewServeMux()
	s.mux.HandleFunc("/index.html", s.handleLogin)
	s.mux.HandleFunc("/auth", s.handleAuth)
	s.mux.HandleFunc("/", s.handleRoot)

	s.waitGroup = &sync.WaitGroup{}
	s.closeChan = make(chan struct{}, 1)
	s.run = true

	for _, l := range s.Listeners {
		s.waitGroup.Add(1)
		go s.serveListener(l)
	}

	// 等所有客户端read循环和服务端accept均正常退出
	s.waitGroup.Wait()
	s.clean()
	s.closeChan <- struct{}{}
	return nil
}

// Shutdown 停止服务
func (s *Server) Shutdown() {
	if s.Listeners != nil {
		s.run = false
		for _, l := range s.Listeners {
			l.Close()
		}
		<-s.closeChan
		close(s.closeChan)
	}
}

func (s *Server) serveListener(l net.Listener) {
	defer s.waitGroup.Done()
	for s.run {
		c, err := l.Accept()
		if err != nil {
			if s.run {
				s.log(err)
			}
			continue
		}
		s.waitGroup.Add(1)
		go s.serveClient(c)
	}
}

func (s *Server) serveClient(c net.Conn) {
	defer s.waitGroup.Done()
	defer c.Close()
	for s.run {
		bufr := bufio.NewReader(c)
		request, err := http.ReadRequest(bufr)
		if err != nil || request == nil {
			if err != io.EOF {
				s.log(err)
			}
			return
		}

		if strings.HasSuffix(request.RequestURI, "/") {
			request.RequestURI += "index.html"
			request.URL.Path += "index.html"
		}

		if request.Method != http.MethodConnect {
			rw := newResponse(request)
			rw.resp.Header.Set("X-Transcend-Version", "1")
			s.mux.ServeHTTP(rw, request)
			if rw.resp.Header.Get("Content-Type") == "" {
				rw.resp.Header.Set("Content-Type", "text/html")
			}
			request.Body.Close()
			err = rw.resp.Write(c)
			if err != nil {
				s.log(err)
				return
			}

			if request.Close || rw.resp.Close {
				return
			}
		} else {
			err = s.handleConnect(request, c)
			request.Body.Close()
			if err != nil && err != io.EOF {
				s.log(err)
			}
			return
		}
	}
}

func (s *Server) clean() {
	// TODO: offline online clients, remove tun device
}

func (s *Server) log(msg interface{}) {
	if s.Loger != nil && msg != nil {
		s.Loger.Println(msg)
	}
}
