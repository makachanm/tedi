package auth

import (
	"context"
	"errors"
	"fmt"
	"html"
	"net"
	"net/http"
	"time"
)

type callbackServer struct {
	srv    *http.Server
	url    string
	result chan callbackResult
}

type callbackResult struct {
	query map[string]string
	err   error
}

type checkFunc func(q map[string]string) (reject string, fatal error)

func startCallbackServer(check checkFunc) (*callbackServer, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}

	cb := &callbackServer{
		url:    fmt.Sprintf("http://%s/callback", ln.Addr().String()),
		result: make(chan callbackResult, 1),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		q := map[string]string{}
		for k, v := range r.URL.Query() {
			if len(v) > 0 {
				q[k] = v[0]
			}
		}

		reject, fatal := check(q)
		if reject != "" {
			http.Error(w, reject, http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if fatal != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, page, "Login failed", html.EscapeString(fatal.Error()))
		} else {
			fmt.Fprintf(w, page, "Logged in", "You can close this tab and go back to the terminal.")
		}

		select {
		case cb.result <- callbackResult{query: q, err: fatal}:
		default:
		}
	})

	cb.srv = &http.Server{Handler: mux, ReadHeaderTimeout: 10 * time.Second}
	go cb.srv.Serve(ln)
	return cb, nil
}

const page = `<!doctype html><meta charset="utf-8"><title>tedi</title>
<body style="font-family:sans-serif;max-width:32em;margin:4em auto"><h1>%s</h1><p>%s</p></body>`

func (cb *callbackServer) wait(ctx context.Context) (map[string]string, error) {
	select {
	case r := <-cb.result:
		return r.query, r.err
	case <-ctx.Done():
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return nil, errors.New("timed out waiting for the browser; try again with -manual")
		}
		return nil, ctx.Err()
	}
}

func (cb *callbackServer) close() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	cb.srv.Shutdown(ctx)
}
