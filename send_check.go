package main

import (
	"goprotobuf.googlecode.com/hg/proto"
	/* "math"*/
	"http"
	"exec"
	"log"
	"os"
	"bufio"
	"bytes"
	/* "strings"*/
	"time"
	"fmt"
)

type Check struct {
	hostname    string
	servicename string
	cmd         []string
	env         []string
	shell       bool
}

// chasing a moving target with exec here. currently it's kind of
// clumsy to pull out the return code. with luck the target will
// move again in a way i like.
func runPlugin(cmd, env []string, timeout int64) (result *CheckResult) {
	var output bytes.Buffer
	rc := 0
	log.Printf("running check %s", cmd)

	/* c := exec.Command(cmd[0], cmd...)*/
	c := exec.Command(cmd[0], cmd[1:]...)
	c.Stdout = &output
	starttime := time.Nanoseconds()
	err := c.Start()
	if err != nil {
		log.Fatal("Error running command ", cmd, ": ", err)
	}
	defer c.Process.Release()
	timer := time.AfterFunc(timeout, func() { c.Process.Kill() })
	err = c.Wait()
	timer.Stop()
	endtime := time.Nanoseconds()
	/* log.Print(msg)*/
	if err != nil {
		if msg, ok := err.(*os.Waitmsg); ok {
			rc = msg.ExitStatus()
		} else {
			log.Print("Error running command ", cmd, ": ", err)
		}
	}

	result = &CheckResult{
		StartTimestamp: proto.Int64(starttime),
		EndTimestamp:   proto.Int64(endtime),
		Status:         NewCheckStatus(CheckStatus(rc)),
		CheckPassive:   proto.Bool(*flagPassive),
	}
	switch rc {
	case 0, 1, 2, 3:
		// this is ok!
		log.Printf("%s: returned %s", cmd, CheckStatus_name[int32(rc)])
		result.Status = NewCheckStatus(CheckStatus(rc))
		result.CheckOutput = proto.String(string(bytes.TrimSpace(output.Bytes())))
		break
	default:
		// XXX check for timeout/sig9, presently assumed
		log.Printf("%s: return code %d", cmd, rc)
		result.Status = NewCheckStatus(CheckStatus_UNKNOWN)
		result.CheckOutput = proto.String(fmt.Sprintf("UNKNOWN: Command timed out after %d seconds\n", *flagCmdTimeout) + string(bytes.TrimSpace(output.Bytes())))
	}
	return result
}

func PostToEndpoint(buf []byte, url string) {
	log.Print("Posting to ", url)
	b := bytes.NewBuffer(buf)

	rq, err := http.NewRequest("POST", url, b)
	if err != nil {
		log.Print("http.NewRequest: ", err)
	}
	rq.Header.Set("Content-Type", "application/x-protobuf")
	if *flagUsername != "" && *flagPassword != "" {
		rq.SetBasicAuth(*flagUsername, *flagPassword)
	}

	resp, err := new(http.Client).Do(rq)
	if err != nil {
		log.Print("Post to endpoint: ", err)
		return
	}
	defer resp.Body.Close()
}

func RunServiceCheck(cmd, env []string, host, service string, shell bool, c chan *CheckResult) {
	var result *CheckResult
	if shell {
		cmd = append([]string{"/bin/sh", "-c"}, cmd...)
	}
	/* log.Printf("running cmd %v", cmd)*/
	result = runPlugin(cmd, nil, *flagCmdTimeout*1e9)
	result.Hostname = proto.String(host)
	result.ServiceName = proto.String(service)
	/* log.Printf("check returned! %s", proto.CompactTextString(result))*/
	c <- result
}

func channelWrap(c chan int, f func(a ...interface{}), args ...interface{}) {
	f(args...)
	c <- 1
}

// reads from a file f with expected format <hostname>,<servicename>,<cmd>\n
// and executes plugin checks concurrently
func runCommandList(f *os.File, msg *CheckResultSet) {
	b := bufio.NewReader(f)
	cmds := 0
	c := make(chan *CheckResult)
	for l, prefix, err := b.ReadLine(); err == nil; l, prefix, err = b.ReadLine() {
		// read until prefix is false
		if prefix {
			for r, prefix, err := b.ReadLine(); prefix && err == nil; r, prefix, err = b.ReadLine() {
				l = append(l, r...)
			}
		}
		line := bytes.SplitN(l, []byte(","), 3)
		if len(line) == 3 {
			host, service, cmd := string(line[0]), string(line[1]), string(line[2])
			go RunServiceCheck([]string{cmd}, nil, host, service, true, c)
			cmds++
		} else {
			log.Print("failed to parse: ", line)
		}
	}
	// drain channel
	for i := 0; i < cmds; i++ {
		msg.Results = append(msg.Results, <-c)
	}
	// split out the hostname and service name
	return
}

func runSingleCheck(msgset *CheckResultSet, cmd []string) {
	c := make(chan *CheckResult)
	go RunServiceCheck(cmd, nil, *flagHostname, *flagServicename, false, c)
	result := <-c
	msgset.Results = append(msgset.Results, result)
}
