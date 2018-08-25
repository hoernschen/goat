package main

import (
	"log"
	"net/http"
	"os"

	"github.com/hoernschen/goat/handler/signaling"
	"github.com/urfave/cli"
)

var keyPath = "./ssl.key"
var certPath = "./ssl.crt"
var htmlPath = "./html/"

func main() {
	app := cli.NewApp()
	app.Name = "Goat"
	app.Usage = "A WebRTC Server"

	app.Commands = []cli.Command{
		{
			Name:    "signaling",
			Aliases: []string{"s"},
			Usage:   "Sets up a Media Router",
			Action: func(c *cli.Context) error {
				log.Println("Start Signaling Server")
				go signaling.Run()
				return nil
			},
		},
		{
			Name:    "router",
			Aliases: []string{"r"},
			Usage:   "Sets up a Media Router",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "type",
					Value: "sfu",
				},
			},
			Action: func(c *cli.Context) error {
				routerType := c.String("type")
				if routerType == "sfu" {
					log.Println("Start Media Router")
					go signaling.RunMediaRouter()
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

	httpErr := http.ListenAndServeTLS(":443", certPath, keyPath, router)
	if httpErr != nil {
		log.Fatal(httpErr)
	}

	go http.ListenAndServe(":80", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://"+r.Host+r.URL.String(), http.StatusMovedPermanently)
	}))
}
