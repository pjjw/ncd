package main

import (
	/* "regexp"*/
	//  "fmt"
	"os"
	"fmt"
	"strconv"
	"strings"
	"time"
	"goprotobuf.googlecode.com/hg/proto"
)

func touchCheckOkFile(f *os.File) (err os.Error) {
	name := f.Name() + ".ok"
	f, err = os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0660)
	if err == nil {
		f.Truncate(0)
		f.Close()
	}
	return
}

func WriteCheck(check *CheckResult, spooldir string) (fn string, err os.Error) {
	f, err := spoolFile(spooldir, "c")
	if err != nil {
		return "", err
	}
	defer f.Close()
	fn = f.Name()
	err = writePerfCheck(f, check)
	if err == nil {
		err = touchCheckOkFile(f)
	}
	return
}

func splitUnits(in string) (val float32, units string, err os.Error) {
	lastnum := strings.LastIndexAny(in, "0123456789.")
	if lastnum == -1 {
		err = os.NewError("non-numeric value")
		return
	}
	if lastnum+1 < len(in) {
		/* pd.Units = proto.String(in[lastnum+1 : len(in)])*/
		units = in[lastnum+1 : len(in)]
		val, err = strconv.Atof32(in[0 : lastnum+1])
		if err != nil {
			return 0, "", err
		}
	} else {
		val, err = strconv.Atof32(in)
	}
	return
}

// parse an individial element of perfdata
func parsePerfDataElement(str string) (pd *PerfData, err os.Error) {
	m := strings.SplitN(str, "=", 2)
	name := m[0]
	pd = &PerfData{Name: proto.String(name)}
	n := strings.Split(m[1], ";")
	/* nf := make([]float32, len(n))*/
	/* for i, v := range n {*/
	/*   nf[i], err = strconv.Atof32(v)*/
	/*   if err != nil {*/
	/*     return*/
	/*   }*/
	/* }*/

	// the below feels very ugly.
	if len(m) < 2 {
		err = os.NewError("no value")
		return nil, err
	}
	// check for units of measurement
	lastnum := strings.LastIndexAny(n[0], "0123456789.")
	if lastnum == -1 {
		err = os.NewError("non-numeric value")
		return
	}
	if lastnum+1 < len(n[0]) {
		pd.Units = proto.String(n[0][lastnum+1 : len(n[0])])
		val, err := strconv.Atof32(n[0][0 : lastnum+1])
		if err != nil {
			return nil, err
		}
		pd.Value = proto.Float32(val)
	} else {
		val, err := strconv.Atof32(n[0])
		if err == nil {
			pd.Value = proto.Float32(val)
		}
	}
	if len(n) < 3 {
		return
	}
	val, err := strconv.Atof32(n[1])
	if err != nil {
		return
	}
	pd.Warning = proto.Float32(val)
	if len(n) < 4 {
		return
	}
	val, err = strconv.Atof32(n[2])
	if err != nil {
		return
	}
	pd.Critical = proto.Float32(val)
	if len(n) < 5 {
		return
	}
	val, err = strconv.Atof32(n[3])
	if err != nil {
		return
	}
	pd.Minimum = proto.Float32(val)
	if len(n) < 6 {
		return
	}
	val, err = strconv.Atof32(n[4])
	if err != nil {
		return
	}
	pd.Maximum = proto.Float32(val)

	return
}

// parse a string containing perfdata
func parsePerfData(in string) (pd []*PerfData, err os.Error) {
	for _, v := range strings.Fields(in) {
		elem, err := parsePerfDataElement(strings.TrimSpace(v))
		if err != nil {
			return nil, err
		}
		pd = append(pd, elem)
	}
	return
}

func ParseRawPluginOutput(in string) (out string, pd []*PerfData, err os.Error) {
	// find first newline and pipe
	firstpipe := strings.Index(in, "|")
	firstcr := strings.Index(in, "\n")
	if firstpipe == -1 {
		// no perfdata, we're done
		out = in
		pd = []*PerfData{}
		return out, pd, nil
	}
	if firstcr == -1 {
		// one-line output
		// XXX for each match parse perf data and append
		pd, err = parsePerfData(in[firstpipe+1 : len(in)])
		out = in[0:firstpipe]
		return out, pd, err
	}
	// multiline output with perfdata
	lastpipe := strings.LastIndex(in, "|")
	out = in[0:firstpipe] + in[firstcr:lastpipe]
	pd, err = parsePerfData(in[firstpipe+1:firstcr] + in[lastpipe+1:len(in)])
	return
}

// write textual representation of perfcheck data to writer
func writePerfCheck(f *os.File, check *CheckResult) (err os.Error) {
	err = tActiveCheck.Execute(f, check.stringMap())
	return
}

func (c *CheckResult) stringMap() (smap map[string]string) {
	smap = map[string]string{
		"Hostname":    fmt.Sprintf("%s", proto.GetString(c.Hostname)),
		"ServiceName": fmt.Sprintf("%s", proto.GetString(c.ServiceName)),
		"Status":      fmt.Sprintf("%d", int32((*c.Status))),
		"CheckPassive": fmt.Sprintf("%d", func() (i int32) {
			if proto.GetBool(c.CheckPassive) {
				i = 1
			} else {
				i = 0
			}
			return i
		}()),
		"CheckOutput":    fmt.Sprintf("%s", strings.Trim(strconv.Quote(proto.GetString(c.CheckOutput)), "\"")),
		"StartTimestamp": fmt.Sprintf("%f", float64(proto.GetInt64(c.StartTimestamp))/1000000000),
		"EndTimestamp":   fmt.Sprintf("%f", float64(proto.GetInt64(c.EndTimestamp))/1000000000),
		"TimeNow":        fmt.Sprintf("%d", time.Seconds()),
	}
	return smap
}
