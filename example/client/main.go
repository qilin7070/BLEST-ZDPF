package main

import (
	"bytes"
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"sync"

	quic "github.com/lucas-clemente/quic-go"

	"github.com/lucas-clemente/quic-go/h2quic"
	"github.com/lucas-clemente/quic-go/internal/utils"
)

func main() {
	println("1")
	verbose := flag.Bool("v", false, "verbose")
	multipath := flag.Bool("m", true, "multipath")
	output := flag.String("o", "", "logging output")
	flag.Parse()
	urls := flag.Args()
	println("2")
	if *verbose {
		utils.SetLogLevel(utils.LogLevelDebug)
	} else {
		utils.SetLogLevel(utils.LogLevelInfo)
	}
	println("3")
	utils.SetLogTimeFormat("")

	if *output != "" {
		logfile, err := os.Create(*output)
		if err != nil {
			panic(err)
		}
		defer logfile.Close()
		log.SetOutput(logfile)
	}
	println("4")
	quicConfig := &quic.Config{
		CreatePaths: *multipath,
	}

	hclient := &http.Client{
		Transport: &h2quic.RoundTripper{QuicConfig: quicConfig},
	}
	println("5")
	var wg sync.WaitGroup
	wg.Add(len(urls))
	for _, addr := range urls {
		println("6")
		utils.Infof("GET %s", addr)
		go func(addr string) {
			println("7")
			rsp, err := hclient.Get(addr)
			println("8")
			if err != nil {
				println("0")
				panic(err)
				println("9")
			}
			utils.Infof("Got response for %s: %#v", addr, rsp)

			body := &bytes.Buffer{}
			_, err = io.Copy(body, rsp.Body)
			if err != nil {
				panic(err)
			}
			utils.Infof("Request Body:")
			utils.Infof("%s", body.Bytes())
			wg.Done()
		}(addr)
	}
	wg.Wait()
}
