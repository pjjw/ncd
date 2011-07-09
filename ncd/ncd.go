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
	addr     = flag.String("addr", ":8323", "http service address")
	endpoint = flag.String("endpoint", "/ncd/", "http service endpoint")
	usessl   = flag.Bool("ssl", false, "use ssl")
	sslcert  = flag.String("cert", "cert.pem", "ssl cert file")
	sslkey   = flag.String("key", "key.pem", "ssl key file")
	spooldir = flag.String("spooldir", "/var/nagios/spool/checkresults", "nagios spool directory")
	debug    = flag.Bool("syslog", true, "log to syslog- if false, log to stdout")
	username = flag.String("username", "npd", "basic auth username")
	password = flag.String("password", "npd", "basic auth password")
	realm    = flag.String("realm", "npd", "basic auth realm")

	logger = syslog.NewLogger(syslog.LOG_INFO, log.Flags())
)


var templ = template.MustParse(templateStr, nil)

func root(w http.ResponseWriter, r *http.Request) {
	// check header
  if *password != "" && *username != "" {
    auth, ok := r.Header["Authorization"]
    if ok && strings.HasPrefix(auth[0], "Basic ") {
      str := strings.TrimLeft(auth[0], "Basic ")
      decode,err := base64.StdEncoding.DecodeString(str)
      if err != nil {
        log.Print("cannot decode auth string: ", err)
        return
      }
      user, pass, err := http.UnescapeUserinfo(string(decode))
      if err != nil {
        log.Print("auth: couldn't decode user/pass: ", err)
      }
      if !(user == *username && pass == *password) {
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
			_, err := WriteCheck(v, *spooldir)
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
	// set up logging
	if !*debug {
		logger = log.New(os.Stderr, log.Prefix(), log.Flags())
	}
	http.HandleFunc(*endpoint, root)
	logger.Print("bringing up endpoint")

  var err os.Error
  if *usessl {
    err = http.ListenAndServeTLS(*addr, *sslcert, *sslkey, nil)
  } else {
    err = http.ListenAndServe(*addr, nil)
  }
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}


const templateStr = `ok`
