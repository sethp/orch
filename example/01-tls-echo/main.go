package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
)

var (
	Config = struct {
		CertFile, KeyFile string
	}{
		CertFile: "/etc/tls/tls.crt",
		KeyFile:  "/etc/tls/tls.key",
	}
)

func init() {
	flag.StringVar(&Config.CertFile, "server-cert", Config.CertFile, "path to certificate file")
	flag.StringVar(&Config.KeyFile, "server-key", Config.KeyFile, "path to key file")
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, r.URL)
	})

	if err := http.ListenAndServeTLS(":8443", Config.CertFile, Config.KeyFile, http.DefaultServeMux); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
