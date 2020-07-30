package server

import (
	"bufio"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"strings"
)

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	writeLoginXML("请输入用户名密码", w)
}

func (s *Server) handleAuth(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	username := r.Form.Get("username")
	password := r.Form.Get("password")

	// TODO: 进行用户名密码验证
	if username != "h" || password != "h" {
		writeLoginXML("用户名密码错误，请重新输入", w)
	} else {
		ret := `<?xml version="1.0" encoding="UTF-8"?>
				<config-auth client="vpn" type="complete">
					<version who="sg">0.1(1)</version>
					<auth id="success">
						<title>SSL VPN Service</title>
					</auth>
				</config-auth>`

		header := w.Header()
		header.Set("Connection", "Keep-Alive")
		header.Set("Content-Type", "text/xml")
		header.Set("Set-Cookie", "webvpncontext=+yXSlV8MpRNnURhSX/+05svIAydLG8ubYAypnKtK2yw=; Secure")
		header.Add("Set-Cookie", "webvpn=+yXSlV8MpRNnURhSX/+05svIAydLG8ubYAypnKtK2yw=; Secure")
		header.Add("Set-Cookie", "webvpnc=; expires=Thu, 01 Jan 1970 22:00:00 GMT; path=/; Secure")
		header.Add("Set-Cookie", "webvpnc=bu:/&p:t&iu:1/&sh:15499E46D4D3E79817D0E341D7A7D87CF0B36760; path=/; Secure")
		w.Write([]byte(ret))
	}
}

func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	upath := r.URL.Path
	if !strings.HasPrefix(upath, "/") {
		upath = "/" + upath
	}

	f, err := s.htmlFS.Open(upath)
	if err != nil {
		s.log(upath)
		writeNotFound(w)
		return
	}
	defer f.Close()
	io.Copy(w, f)

	extPos := strings.LastIndexByte(upath, '.')
	if extPos == -1 {
		return
	}
	ext := upath[extPos:]
	if ext == "." {
		return
	}
	ct := mime.TypeByExtension(ext)
	if ct != "" {
		charsetPos := strings.IndexByte(ct, ';')
		if charsetPos != -1 {
			ct = strings.TrimSpace(ct[:charsetPos])
		}
		w.Header().Set("Content-Type", ct)
	}
}

func (s *Server) handleConnect(r *http.Request, client net.Conn) error {
	header := r.Header

	w := bufio.NewWriter(client)

	w.WriteString("HTTP/1.1 200 CONNECTED\r\n")
	w.WriteString("X-CSTP-Version: 1\r\n")
	w.WriteString("X-CSTP-Server-Name: anyconnect-server-go 1.0.1\r\n")
	fmt.Fprintf(w, "X-CSTP-Hostname: %v\r\n", header.Get("X-CSTP-Hostname"))
	w.WriteString("X-CSTP-DPD: 1800\r\n")
	w.WriteString("X-CSTP-Address: 192.168.35.222\r\n")
	w.WriteString("X-CSTP-Netmask: 255.255.240.0\r\n")
	w.WriteString("X-CSTP-DNS: 119.29.29.29\r\n")
	w.WriteString("X-CSTP-Split-Include-IP6: 2000::/3\r\n")
	w.WriteString("X-CSTP-Tunnel-All-DNS: false\r\n")
	w.WriteString("X-CSTP-Split-Exclude: 10.0.0.0/255.0.0.0\r\n")
	w.WriteString("X-CSTP-Split-Exclude: 192.168.0.0/255.255.0.0\r\n")
	w.WriteString("X-CSTP-Split-Exclude: 192.168.0.0/255.255.255.0\r\n")
	w.WriteString("X-CSTP-Split-Exclude: 172.16.0.0/255.240.0.0\r\n")
	w.WriteString("X-CSTP-Split-Exclude: 100.64.0.0/255.192.0.0\r\n")
	w.WriteString("X-CSTP-Split-Exclude: 224.0.0.0/255.0.0.0\r\n")
	w.WriteString("X-CSTP-Split-Exclude: 169.254.0.0/255.255.0.0\r\n")
	w.WriteString("X-CSTP-Keepalive: 32400\r\n")
	w.WriteString("X-CSTP-Idle-Timeout: none\r\n")
	w.WriteString("X-CSTP-Smartcard-Removal-Disconnect: true\r\n")
	w.WriteString("X-CSTP-DynDNS: true\r\n")
	w.WriteString("X-CSTP-Rekey-Time: 172782\r\n")
	w.WriteString("X-CSTP-Rekey-Method: ssl\r\n")
	w.WriteString("X-CSTP-Session-Timeout: none\r\n")
	w.WriteString("X-CSTP-Disconnected-Timeout: none\r\n")
	w.WriteString("X-CSTP-Keep: true\r\n")
	w.WriteString("X-CSTP-TCP-Keepalive: true\r\n")
	w.WriteString("X-CSTP-License: accept\r\n")
	w.WriteString("X-DTLS-DPD: 1800\r\n")
	w.WriteString("X-DTLS-Port: 1443\r\n")
	w.WriteString("X-DTLS-Rekey-Time: 172792\r\n")
	w.WriteString("X-DTLS-Rekey-Method: ssl\r\n")
	w.WriteString("X-DTLS-Keepalive: 32400\r\n")
	w.WriteString("X-DTLS-Session-ID: 459a879b197f41b97646541b2d43e3130ddd4d88c82448f2f93ca09547083729\r\n")
	w.WriteString("X-DTLS12-CipherSuite: AES256-GCM-SHA384\r\n")
	w.WriteString("X-DTLS-MTU: 1434\r\n")
	w.WriteString("X-CSTP-Base-MTU: 1500\r\n")
	w.WriteString("X-CSTP-MTU: 1434\r\n")
	w.WriteString("X-DTLS-Content-Encoding: lzs\r\n")
	w.WriteString("X-CSTP-Content-Encoding: lzs\r\n")
	w.WriteString("\r\n")
	w.Flush()

	// TODO: 调用外部指令
	// TODO: warp tun设备

	fmt.Println("tunnel connected")
	for s.run {
		buff := make([]byte, 2048)
		n, err := client.Read(buff)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		fmt.Println(string(buff[:n]))
	}
	return nil
}

func writeNotFound(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
	header := w.Header()
	header.Set("Connection", "close")
	w.Write([]byte("<html><body><h1>404 Not Found</h1></body></html>"))
}

func writeLoginXML(msg string, w http.ResponseWriter) {
	tpl := `<?xml version="1.0" encoding="UTF-8"?>
	<config-auth client="vpn" type="auth-request">
		<version who="sg">0.1(1)</version>
		<auth id="main">
			<message>%v</message>
			<form method="post" action="/auth">
				<input type="text" name="username" label="用户名:" />
				<input type="password" name="password" label="密码:" />
			</form>
		</auth>
	</config-auth>`

	header := w.Header()
	header.Set("Content-Type", "text/xml")
	header.Set("Set-Cookie", "webvpncontext=; expires=Thu, 01 Jan 1970 22:00:00 GMT; path=/; Secure")

	fmt.Fprintf(w, tpl, msg)
}
