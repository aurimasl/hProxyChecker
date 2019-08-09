package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/ip2location/ip2proxy-go"
)

var (
	port     = flag.String("port", "443", "Listen port")
	certPath = flag.String("cert", "/etc/pki/tls/certs/hostinger.crt", "Certificate path")
	keyPath  = flag.String("key", "/etc/pki/tls/private/hostinger.key", "Private key path")
	dbFile   = flag.String("db", "./IP2PROXY-IP-PROXYTYPE-COUNTRY.BIN", "Path to IP2PROXY db file")
	header   = flag.String("header", "", "Secret header")
)

func init() {
	flag.Parse()
	fmt.Printf("Starting on port %s with secret header '%s'\n", *port, *header)
}

type result struct {
	ModuleVersion   string `json:"moduleVersion"`
	PackageVersion  string `json:"packageVersion"`
	DatabaseVersion string `json:"databaseVersion"`
	ProxyType       string `json:"proxyType"`
	IsProxy         int8   `json:"isProxy"`
}

func isProxy(ip string) (result, error) {
	var res result
	if ip2proxy.Open(*dbFile) == 0 {

		res.ModuleVersion = ip2proxy.ModuleVersion()
		res.PackageVersion = ip2proxy.PackageVersion()
		res.DatabaseVersion = ip2proxy.DatabaseVersion()

		res.IsProxy = ip2proxy.IsProxy(ip)
		res.ProxyType = ip2proxy.GetProxyType(ip)

		defer ip2proxy.Close()

		return res, nil
	}
	fmt.Printf("Error reading BIN file.\n")

	return res, fmt.Errorf("error")
}

func handler(w http.ResponseWriter, r *http.Request) {
	h := r.Header.Get("secret")
	if h != *header {
		http.Error(w, "404 page not found", http.StatusNotFound)
		return
	}
	ip := r.URL.Path[len("/checkproxy/"):]
	result, _ := isProxy(ip)
	jsonResult, _ := json.Marshal(result)
	w.Write([]byte(jsonResult))
}

func main() {
	http.HandleFunc("/checkproxy/", handler)
	if *port == "443" {
		log.Fatal(http.ListenAndServeTLS(":443", *certPath, *keyPath, nil))
	} else {
		log.Fatal(http.ListenAndServe(":"+*port, nil))
	}
}
