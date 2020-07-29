package server

import (
	"bufio"
	"crypto/tls"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Server anyconnect server
type Server struct {
	HTMLFileSystem http.FileSystem
	Loger          *log.Logger
	Timeout        time.Duration

	l   net.Listener
	run bool
	wg  *sync.WaitGroup
	mux *http.ServeMux
}

// ListenAndServeTLS 开始工作
func (s *Server) ListenAndServeTLS(network, addr string, config *tls.Config) error {

	var err error

	s.l, err = tls.Listen(network, addr, config)
	if err != nil {
		return err
	}

	s.mux = http.NewServeMux()
	s.mux.HandleFunc("/auth", s.handleAuth)
	s.mux.HandleFunc("/index.html", s.handleLogin)
	s.mux.HandleFunc("/", s.handleRoot)

	s.run = true
	s.wg = &sync.WaitGroup{}
	for s.run {
		c, err := s.l.Accept()
		if err != nil {
			s.log(err)
			continue
		}

		s.wg.Add(1)
		go s.serve(c)
	}

	s.wg.Wait()
	return nil
}

func (s *Server) serve(c net.Conn) {
	defer s.wg.Done()
	defer c.Close()
	for s.run {
		if s.Timeout > 0 {
			c.SetDeadline(time.Now().Add(s.Timeout))
		}
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

func (s *Server) log(msg interface{}) {
	if s.Loger != nil && msg != nil {
		s.Loger.Println(msg)
	}
}
