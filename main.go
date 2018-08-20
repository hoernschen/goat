package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/hoernschen/goat/handler/signaling"
	"github.com/urfave/cli"
)

//var keyPath = "C:/SSL/ssl.key"
var keyPath = "/home/hoernschen/ssl.key"

//var certPath = "C:/SSL/ssl.crt"
var certPath = "/home/hoernschen/ssl.crt"
var htmlPath = "./html/"

func main() {
	app := cli.NewApp()
	app.Name = "Goat"
	app.Usage = "A WebRTC Server"

	Flags := []cli.Flag{
		cli.StringFlag{
			Name:  "host",
			Value: "tutorialedge.net",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:  "ns",
			Usage: "Looks up the Name Servers for a Particular Host",
			Flags: Flags,
			Action: func(c *cli.Context) error {
				ns, err := net.LookupNS(c.String("host"))
				if err != nil {
					return err
				}
				for i := 0; i < len(ns); i++ {
					fmt.Println(ns[i].Host)
				}
				return nil
			},
		},
		{
			Name:  "ip",
			Usage: "Looks up the IP addresses for a particular host",
			Flags: Flags,
			Action: func(c *cli.Context) error {
				ip, err := net.LookupIP(c.String("host"))
				if err != nil {
					fmt.Println(err)
					return err
				}
				for i := 0; i < len(ip); i++ {
					fmt.Println(ip[i])
				}
				return nil
			},
		},
		{
			Name:  "cname",
			Usage: "Looks up the CNAME for a particular host",
			Flags: Flags,
			Action: func(c *cli.Context) error {
				cname, err := net.LookupCNAME(c.String("host"))
				if err != nil {
					fmt.Println(err)
					return err
				}
				fmt.Println(cname)
				return nil
			},
		},
		{
			Name:  "mx",
			Usage: "Looks up the MX records for a particular host",
			Flags: Flags,
			Action: func(c *cli.Context) error {
				mx, err := net.LookupMX(c.String("host"))
				if err != nil {
					fmt.Println(err)
					return err
				}
				for i := 0; i < len(mx); i++ {
					fmt.Println(mx[i].Host, mx[i].Pref)
				}
				return nil
			},
		},
	}

	appErr := app.Run(os.Args)
	if appErr != nil {
		log.Fatal(appErr)
	}

	router := NewRouter()

	router.PathPrefix("/").Handler(http.FileServer(http.Dir(htmlPath)))
	//router.Methods("GET").Path("/ws").Name("Websocket Connection").HandlerFunc(handler.WSConnection)
	//go handler.WSMessages()
	//go signaling.Run()
	go signaling.RunMediaServer()

	httpErr := http.ListenAndServeTLS(":443", certPath, keyPath, router)
	if httpErr != nil {
		log.Fatal(httpErr)
	}

	go http.ListenAndServe(":80", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://"+r.Host+r.URL.String(), http.StatusMovedPermanently)
	}))
}
