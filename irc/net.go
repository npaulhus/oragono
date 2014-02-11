package irc

import (
	"bufio"
	"io"
	"log"
	"net"
	"strings"
)

// Adapt `net.Conn` to a `chan string`.
func StringReadChan(conn net.Conn) <-chan string {
	ch := make(chan string)
	reader := bufio.NewReader(conn)
	go func() {
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err != io.EOF {
					log.Printf("%s → %s error: %s", conn.RemoteAddr(), conn.LocalAddr(), err)
				}
				break
			}
			if DEBUG_NET {
				log.Printf("%s → %s %s", conn.RemoteAddr(), conn.LocalAddr(), line)
			}

			ch <- strings.TrimSpace(line)
		}
	}()
	return ch
}

func maybeLogWriteError(conn net.Conn, err error) bool {
	if err != nil {
		if err != io.EOF {
			log.Printf("%s ← %s error: %s", conn.RemoteAddr(), conn.LocalAddr(), err)
		}
		return true
	}
	return false
}

func StringWriteChan(conn net.Conn) chan<- string {
	ch := make(chan string)
	writer := bufio.NewWriter(conn)
	go func() {
		for str := range ch {
			if DEBUG_NET {
				log.Printf("%s ← %s %s", conn.RemoteAddr(), conn.LocalAddr(), str)
			}
			if _, err := writer.WriteString(str); maybeLogWriteError(conn, err) {
				break
			}
			if _, err := writer.WriteString(CRLF); maybeLogWriteError(conn, err) {
				break
			}
			if err := writer.Flush(); maybeLogWriteError(conn, err) {
				break
			}
		}
	}()
	return ch
}

func AddrLookupHostname(addr net.Addr) string {
	addrStr := addr.String()
	ipaddr, _, err := net.SplitHostPort(addrStr)
	if err != nil {
		return addrStr
	}
	return LookupHostname(ipaddr)
}

func LookupHostname(addr string) string {
	names, err := net.LookupAddr(addr)
	if err != nil {
		return addr
	}
	return strings.TrimSuffix(names[0], ".")
}