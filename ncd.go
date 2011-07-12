package main

import (
	"log"
	"flag"
	"http"
	"os"
	"strings"
	"encoding/base64"
	"bytes"
	"json"
	"syslog"
	"template"
	"goprotobuf.googlecode.com/hg/proto"
)

var (
	flagServer      = flag.Bool("server", false, "run ncd server")
	flagCmdTimeout  = flag.Int64("timeout", 10, "check command timeout (in secs)")
	flagHostname    = flag.String("host", "", "reported hostname")
	flagServicename = flag.String("service", "", "reported servicename")
	flagURL         = flag.String("url", "http://127.0.0.1:8323/ncd/", "url endpoint to post check data to")
	flagSilent      = flag.Bool("silent", false, "suppress output")
	flagPassive     = flag.Bool("passive", true, "submit as passive check")
	flagCmdlist     = flag.Bool("cmdlist", false, "arg is a file containing commands")
	flagAddr        = flag.String("addr", ":8323", "http service address")
	flagEndpoint    = flag.String("endpoint", "/ncd/", "http service endpoint")
	flagUseSSL      = flag.Bool("ssl", false, "use ssl")
	flagSSLCert     = flag.String("cert", "cert.pem", "ssl cert file")
	flagSSLKey      = flag.String("key", "key.pem", "ssl key file")
	flagSpoolDir    = flag.String("spooldir", "/var/nagios/spool/checkresults", "nagios spool directory")
	flagSyslog      = flag.Bool("syslog", true, "log to syslog- if false, log to stdout")
	flagUsername    = flag.String("username", "npd", "basic auth username")
	flagPassword    = flag.String("password", "npd", "basic auth password")

	logger = syslog.NewLogger(syslog.LOG_INFO, log.Flags())
)


var templ = template.MustParse(templateStr, nil)

func root(w http.ResponseWriter, r *http.Request) {
	// check header
	if *flagPassword != "" && *flagUsername != "" {
		auth, ok := r.Header["Authorization"]
		if ok && strings.HasPrefix(auth[0], "Basic ") {
			str := strings.TrimLeft(auth[0], "Basic ")
			decode, err := base64.StdEncoding.DecodeString(str)
			if err != nil {
				log.Print("cannot decode auth string: ", err)
				return
			}
			user, pass, err := http.UnescapeUserinfo(string(decode))
			if err != nil {
				log.Print("auth: couldn't decode user/pass: ", err)
			}
			if !(user == *flagUsername && pass == *flagPassword) {
				log.Print("auth: wrong user/pass: ", user+"/"+pass, *r)
				return
			}
			/* log.Printf("auth: %#v, user: %s, pass: %s", auth, user, pass)*/
		} else {
			log.Print("auth: no authorization")
			return
		}
	}

	checkpb := new(CheckResultSet)
	if r.Method == "POST" {
		cout := new(bytes.Buffer)
		if _, err := cout.ReadFrom(r.Body); err != nil {
			log.Print("error! ", err)
			return
		}
		switch r.Header["Content-Type"][0] {
		case "application/x-protobuf":
			err := proto.Unmarshal(cout.Bytes(), checkpb)
			if err != nil {
				log.Printf("unmarshalling error: ", err)
			}
		case "application/json", "text/plain", "application/x-www-form-urlencoded", "multipart/form-data":
			err := json.Unmarshal(cout.Bytes(), checkpb)
			if err != nil {
				log.Printf("unmarshalling error: ", err)
			}
		}
		logger.Printf("check returned! %s", proto.CompactTextString(checkpb))
		for _, v := range checkpb.Results {
			_, err := WriteCheck(v, *flagSpoolDir)
			if err != nil {
				logger.Print("writecheck failed: ", err)
			}
		}
	} else {
		/* logger.Printf("NOT POST!! %s", r.Method)*/
	}
	templ.Execute(w, nil)
}

func main() {
	flag.Parse()
	cmd := flag.Args()
	// set up logging
	if !*flagSyslog {
		logger = log.New(os.Stderr, log.Prefix(), log.Flags())
	}
	switch *flagServer {
	case true:

		http.HandleFunc(*flagEndpoint, root)
		logger.Print("bringing up endpoint")

		var err os.Error
		if *flagUseSSL {
			err = http.ListenAndServeTLS(*flagAddr, *flagSSLCert, *flagSSLKey, nil)
		} else {
			err = http.ListenAndServe(*flagAddr, nil)
		}
		if err != nil {
			log.Fatal("ListenAndServe:", err)
		}
	case false:
		if *flagSilent {
			devnull, _ := os.Open(os.DevNull)
			log.SetOutput(devnull)
		}

		switch {
		case *flagHostname == "":
			log.Fatal("missing hostname")
		case *flagServicename == "":
			log.Fatal("missing servicename")
		case len(cmd) < 1:
			log.Fatal("missing command(s)")
		}

		msg := new(CheckResultSet)

		if *flagCmdlist {
			for _, v := range cmd {
				f, err := os.Open(v)
				if err != nil {
					log.Print("error: ", err)
				} else {
					defer f.Close()
					runCommandList(f, msg)
				}
			}
		} else {
			runSingleCheck(msg, cmd)
		}

		buf, err := proto.Marshal(msg)
		if err != nil {
			log.Fatal("marshalling error: ", err)
		}

		PostToEndpoint(buf, *flagURL)
	}
}


const templateStr = `ok`
