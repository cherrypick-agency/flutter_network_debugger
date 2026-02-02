package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	socket "github.com/zishang520/socket.io/v2/socket"
)

func main() {
	var addr string
	var socketPath string
	flag.StringVar(&addr, "addr", "127.0.0.1:0", "listen address")
	flag.StringVar(&socketPath, "path", "/socket.io/", "Socket.IO path (e.g. /socket.io/)")
	flag.Parse()

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	io := socket.NewServer(nil, nil)
	io.On("connection", func(clients ...any) {
		client := clients[0].(*socket.Socket)
		client.On("hello", func(datas ...any) {
			client.Emit("hello", datas...)
		})
	})
	// Extra namespace for e2e tests.
	io.Of("/chat", nil).On("connect", func(clients ...any) {
		client := clients[0].(*socket.Socket)
		client.On("hello", func(datas ...any) {
			client.Emit("hello", datas...)
		})
	})

	p := strings.TrimSpace(socketPath)
	if p == "" {
		p = "/socket.io/"
	}
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	if !strings.HasSuffix(p, "/") {
		p = p + "/"
	}

	mux := http.NewServeMux()
	mux.Handle(p, io.ServeHandler(nil))
	// Redirect the non-slashed path to the canonical one.
	mux.HandleFunc(strings.TrimSuffix(p, "/"), func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, p, http.StatusPermanentRedirect)
	})
	srv := &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	fmt.Printf("listening %s\n", ln.Addr().String())

	go func() {
		if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
}
