// Copyright 2013 v.chu.pe authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"github.com/fiorix/go-redis/redis"
	"github.com/fiorix/go-web/httpxtra"
	_ "github.com/go-sql-driver/mysql"
	"html/template"
	"log"
	"net/http"
	"runtime"
	"sync"
	"time"
)

const (
	VERSION = "1.0"
	APPNAME = "v.chu.pe"
)

var (
	Config *ConfigData
	Redis  *redis.Client
	Room   *template.Template
	Status *template.Template
)

func route() {
	// Public handlers: add your own
	http.Handle("/", http.FileServer(http.Dir(Config.DocumentRoot)))
	http.HandleFunc("/r/", RoomHandler)
	http.HandleFunc("/s/", RoomStatusHandler)
	http.HandleFunc("/new", NewRoomHandler)
}

func hello() {
	var cpuinfo string
	if n := runtime.NumCPU(); n > 1 {
		runtime.GOMAXPROCS(n)
		cpuinfo = fmt.Sprintf("%d CPUs", n)
	} else {
		cpuinfo = "1 CPU"
	}
	log.Printf("%s v%s (%s)", APPNAME, VERSION, cpuinfo)
}

func main() {
	var err error
	cfgfile := flag.String("config", "server.conf", "set config file")
	flag.Parse()
	Config, err = ReadConfig(*cfgfile)
	if err != nil {
		log.Fatal(err)
	}

	Room = template.Must(template.ParseFiles(Config.Templates + "/room.html"))
	Status = template.Must(template.ParseFiles(Config.Templates + "/status.html"))

	// Set up databases
	log.Println("Setup databases")
	Redis = redis.New(Config.Redis)

	// Set up routing and print server info
	log.Println("Template folder: " + Config.Templates)
	route()
	hello()
	// Run HTTP and HTTPS servers
	wg := &sync.WaitGroup{}
	if Config.HTTP.Addr != "" {
		wg.Add(1)
		log.Printf("Starting HTTP server on %s", Config.HTTP.Addr)
		go func() {
			// Use httpxtra's listener to support Unix sockets.
			server := http.Server{
				Addr: Config.HTTP.Addr,
				Handler: httpxtra.Handler{
					Logger:   logger,
					XHeaders: Config.HTTP.XHeaders,
				},
			}
			log.Fatal(httpxtra.ListenAndServe(server))
			//wg.Done()
		}()
	}
	if Config.HTTPS.Addr != "" {
		wg.Add(1)
		log.Printf("Starting HTTPS server on %s", Config.HTTPS.Addr)
		go func() {
			server := http.Server{
				Addr:    Config.HTTPS.Addr,
				Handler: httpxtra.Handler{Logger: logger},
			}
			log.Fatal(server.ListenAndServeTLS(
				Config.HTTPS.CrtFile, Config.HTTPS.KeyFile))
			//wg.Done()
		}()
	}
	wg.Wait()
}

func logger(r *http.Request, created time.Time, status, bytes int) {
	//fmt.Println(httpxtra.ApacheCommonLog(r, created, status, bytes))
	log.Printf("HTTP %d %s %s (%s) :: %s",
		status,
		r.Method,
		r.URL.Path,
		r.RemoteAddr,
		time.Since(created))
}
