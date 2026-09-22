package internal

import (
	"encoding/json"
	goc "gor/pkg/client"
	gor "gor/pkg/server"
	"io"
	"log/slog"
	"net/http"
	"time"
)

type Event struct {
	ID  int
	Seq int
}

func RunFullDuplex() {
	go time.AfterFunc(1*time.Second, RunClient)
	RunServer()
}

func RunServer() {

	opts := gor.NewServerOpts().
		WithReadHeaderTimeout(0).
		WithReadTimeout(0).
		WithWriteTimeout(0).
		WithIdleTimeout(0)

	g := gor.NewEngine(opts)
	r := g.NewRouter("")
	r.HandleHTTPFunc("GET /stream", func(w http.ResponseWriter, r *http.Request) {
		// https://www.youtube.com/watch?v=-lcH3qrkh_U&t=5s
		// w.Header().Set("Transfer-Encoding", "chunked")
		// w.Header().Set("Content-Type", "application/x-ndjson; charset=utf-8")
		w.Header().Set("Content-Type", "application/json; charset=utf-8")

		channel := make(chan int, 1)
		c := http.NewResponseController(w)
		// IMPORTANT!
		if err := c.EnableFullDuplex(); err != nil {
			panic("server: full duplext not supported " + err.Error())
		}
		d := json.NewDecoder(r.Body)
		go func() {
			for {
				if r.Context().Err() != nil {
					break
				}
				var event struct {
					Next int
				}

				err := d.Decode(&event)

				if err != nil {
					slog.Error("server: error on event", "error", err)
					break
				} else {
					slog.Info("server: on event", "event", event)
					channel <- event.Next
				}
			}
		}()

		e := json.NewEncoder(w)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(206)
		epoch := 0
		for n := range channel {
			slog.Info("server: handling request", "id", epoch, "next", n)
			for i := range n {
				err := e.Encode(Event{
					ID:  epoch,
					Seq: i + 1,
				})
				if err != nil {
					slog.Error("server: send response error", "error", err)
				}
				if err := c.Flush(); err != nil {
					slog.Error("server: flush error", "error", err)
				}

				time.Sleep(time.Second)
			}
			epoch++
		}
	})

	if err := g.Listen(":8080"); err != nil {
		slog.Error("app stopped with an error", "error", err)
	}

}

func RunClient() {
	client := goc.NewClient(goc.NewClientOpts().
		WithTimeout(0))
	ch := make(chan struct{})
	r, w := io.Pipe()
	go func() {
		defer w.Close()
		e := json.NewEncoder(w)
		for i := range 1000 {
			req := map[string]int{
				"next": i,
			}
			err := e.Encode(req)
			if err != nil {
				slog.Error("client: can not send request", "error", err, "next", i)
				return
			} else {
				slog.Info("client: request more", "next", i)
			}
			for i > 0 {
				<-ch
				i--
			}
		}
	}()
	req, _ := http.NewRequest("GET", "http://localhost:8080/stream", r)
	// Do() returns as soon as the server sends headers, even if the body is still streaming!
	resp, err := client.Do(req)
	slog.Info("client: after do")
	if err != nil {
		slog.Error("client: error", "error", err)
	}
	defer resp.Body.Close()
	d := json.NewDecoder(resp.Body)

	for {
		var event Event
		err := d.Decode(&event)
		if err != nil {
			slog.Error("client: decode error", "error", err)
			return
		}
		slog.Info("client: receive event", "event", event)
		ch <- struct{}{}
	}
}
