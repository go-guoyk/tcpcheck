package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"github.com/google/uuid"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"time"
)

var (
	optDaemon bool
	optAddr   string
	optSource string
	optReport string

	optTCPAddr *net.TCPAddr
)

func main() {
	var err error
	defer func(err *error) {
		if *err != nil {
			log.Println("exited with error:", (*err).Error())
			os.Exit(1)
		}
	}(&err)

	flag.BoolVar(&optDaemon, "daemon", false, "run in server mode")
	flag.StringVar(&optAddr, "addr", "127.0.0.1:7111", "bind address for server mode or target address for client mode")
	flag.StringVar(&optSource, "source", "unknown", "source name for client mode")
	flag.StringVar(&optReport, "report", "http://127.0.0.1:9200/tcpcheck/_doc", "report url to post json")
	flag.Parse()

	if optTCPAddr, err = net.ResolveTCPAddr("tcp", optAddr); err != nil {
		return
	}

	if optDaemon {
		var l *net.TCPListener
		if l, err = net.ListenTCP("tcp", optTCPAddr); err != nil {
			return
		}
		for {
			var conn *net.TCPConn
			if conn, err = l.AcceptTCP(); err != nil {
				return
			}
			rand.Seed(time.Now().UnixNano())
			go serverRoutine(conn)
		}
	} else {
		go clientRoutineShort()
		go clientRoutineLong()
		<-make(chan struct{})
	}

}

func serverRoutine(conn *net.TCPConn) {
	var err error
	defer conn.Close()

	log.Printf("new connection: %s", conn.RemoteAddr().String())

	conn.SetNoDelay(true)
	conn.SetKeepAlive(true)

	rBuf := make([]byte, 4, 4)
	wBuf := make([]byte, 1000000, 1000000)
	for {
		if _, err = conn.Read(rBuf); err != nil {
			log.Printf("%s failed to read", conn.RemoteAddr().String())
			return
		}

		switch string(rBuf) {
		case "RDTR":
			if _, err = conn.Write([]byte("OK")); err != nil {
				log.Printf("%s failed to write", conn.RemoteAddr().String())
				return
			}
		case "T10M":
			rand.Read(wBuf)
			if _, err = io.Copy(conn, bytes.NewReader(wBuf)); err != nil {
				log.Printf("%s failed to write 10m", conn.RemoteAddr().String())
				return
			}
			return
		default:
			log.Printf("%s unknown command: %s", conn.RemoteAddr().String(), rBuf)
			return
		}
	}
}

func clientSubmitRecord(record Record) {
	go func() {
		var err error
		var data []byte
		if data, err = json.Marshal(record); err != nil {
			log.Println("failed to marshal report:", err.Error())
			return
		}
		var resp *http.Response
		if resp, err = http.Post(optReport, "application/json", bytes.NewReader(data)); err != nil {
			log.Println("failed to post report:", err.Error())
			return
		}
		defer resp.Body.Close()
	}()
}

func clientRoutineLong() {
	for {
		clientCheckLong()
		time.Sleep(time.Second * 10)
	}
}

func clientCheckLong() {
	r := Record{
		Source:         optSource,
		Destination:    optTCPAddr.String(),
		ConnectionType: ConnectionLong,
		ConnectionID:   uuid.NewString(),
	}

	var err error

	sw := NewStopWatch()
	var conn *net.TCPConn
	if conn, err = net.DialTCP("tcp", nil, optTCPAddr); err != nil {
		clientSubmitRecord(r.CloneFailure(
			ActionConnect,
			sw.Stop(),
			err.Error(),
		))
		return
	}
	defer conn.Close()

	_ = conn.SetKeepAlive(true)
	_ = conn.SetNoDelay(true)

	clientSubmitRecord(r.CloneSuccess(
		ActionConnect,
		sw.Stop(),
	))

	for {
		sw.Reset()
		if _, err = conn.Write([]byte("RDTR")); err != nil {
			clientSubmitRecord(r.CloneFailure(
				ActionRoundTrip,
				sw.Stop(),
				err.Error(),
			))
			return
		}
		buf := make([]byte, 2, 2)
		if _, err = conn.Read(buf); err != nil {
			clientSubmitRecord(r.CloneFailure(
				ActionRoundTrip,
				sw.Stop(),
				err.Error(),
			))
			return
		}
		if string(buf) != "OK" {
			clientSubmitRecord(r.CloneFailure(
				ActionRoundTrip,
				sw.Stop(),
				"server did not respond OK",
			))
			return
		}
		clientSubmitRecord(r.CloneSuccess(
			ActionRoundTrip,
			sw.Stop(),
		))

		time.Sleep(time.Second * 10)
	}
}

func clientRoutineShort() {
	for {
		clientCheckShort()
		time.Sleep(time.Second * 30)
	}
}

func clientCheckShort() {
	r := Record{
		Source:         optSource,
		Destination:    optTCPAddr.String(),
		ConnectionType: ConnectionShort,
		ConnectionID:   uuid.NewString(),
	}

	var err error

	sw := NewStopWatch()
	var conn *net.TCPConn
	if conn, err = net.DialTCP("tcp", nil, optTCPAddr); err != nil {
		clientSubmitRecord(r.CloneFailure(
			ActionConnect,
			sw.Stop(),
			err.Error(),
		))
		return
	}
	defer conn.Close()

	_ = conn.SetKeepAlive(true)
	_ = conn.SetNoDelay(true)

	clientSubmitRecord(r.CloneSuccess(
		ActionConnect,
		sw.Stop(),
	))

	sw.Reset()
	if _, err = conn.Write([]byte("RDTR")); err != nil {
		clientSubmitRecord(r.CloneFailure(
			ActionRoundTrip,
			sw.Stop(),
			err.Error(),
		))
		return
	}
	buf := make([]byte, 2, 2)
	if _, err = conn.Read(buf); err != nil {
		clientSubmitRecord(r.CloneFailure(
			ActionRoundTrip,
			sw.Stop(),
			err.Error(),
		))
		return
	}
	if string(buf) != "OK" {
		clientSubmitRecord(r.CloneFailure(
			ActionRoundTrip,
			sw.Stop(),
			"server did not respond OK",
		))
		return
	}
	clientSubmitRecord(r.CloneSuccess(
		ActionRoundTrip,
		sw.Stop(),
	))

	sw.Reset()
	if _, err = conn.Write([]byte("T10M")); err != nil {
		clientSubmitRecord(r.CloneFailure(
			ActionTransfer10m,
			sw.Stop(),
			err.Error(),
		))
		return
	}
	var n int64
	if n, err = io.Copy(ioutil.Discard, conn); err != nil {
		clientSubmitRecord(r.CloneFailure(
			ActionTransfer10m,
			sw.Stop(),
			err.Error(),
		))
		return
	}
	if n != 1000000 {
		clientSubmitRecord(r.CloneFailure(
			ActionTransfer10m,
			sw.Stop(),
			"server did not respond 10m bytes",
		))
		return
	}
	clientSubmitRecord(r.CloneSuccess(
		ActionTransfer10m,
		sw.Stop(),
	))
	return
}
