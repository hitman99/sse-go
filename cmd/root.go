package cmd

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/hitman99/sse-go/internal/broker"
	"github.com/spf13/cobra"
	"log"
	"net/http"
	"os"
	"time"
)

var (
	timeout time.Duration
	host    string
	port    string
)

var rootCmd = &cobra.Command{
	Use: "sse-server",
	Run: func(cmd *cobra.Command, args []string) {
		server()
		os.Exit(0)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().DurationVar(&timeout, "timeout", time.Second*30, "client timeout duration")
	rootCmd.PersistentFlags().StringVar(&host, "host", "0.0.0.0", "web server host")
	rootCmd.PersistentFlags().StringVar(&port, "port", "8080", "port for incoming connections")
}

func server() {
	b := broker.NewBroker(timeout)
	r := mux.NewRouter()
	r.HandleFunc("/{topic}", b.RegisterSubscriber()).Methods("GET")
	r.HandleFunc("/{topic}", b.Notify()).Methods("POST")

	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%s", host, port),
		Handler: r,
	}
	log.Printf("started http server on %s", srv.Addr)
	log.Fatal(srv.ListenAndServe())
}
