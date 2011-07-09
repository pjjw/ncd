package main

import (
//  "regexp"
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
  if err == nil {
    defer f.Close()
    fn = f.Name()
    err = writePerfCheck(f, check)
    if err == nil {
      err = touchCheckOkFile(f)
    }
  }
  return
}

// write textual representation of perfcheck data to writer
func writePerfCheck(f *os.File, check *CheckResult) (err os.Error) {
  err = tActiveCheck.Execute(f, check.stringMap())
  return
}

func (c *CheckResult) stringMap() (smap map[string]string) {
  smap = map[string] string {
          "Hostname": fmt.Sprintf("%s", proto.GetString(c.Hostname)),
          "ServiceName": fmt.Sprintf("%s", proto.GetString(c.ServiceName)),
          "Status": fmt.Sprintf("%d", int32((*c.Status))),
          "CheckPassive": fmt.Sprintf("%d", func() (i int32) {if proto.GetBool(c.CheckPassive) {i = 1} else {i = 0}; return i}()),
          "CheckOutput": fmt.Sprintf("%s", strings.Trim(strconv.Quote(proto.GetString(c.CheckOutput)), "\"")),
          "StartTimestamp": fmt.Sprintf("%f", float64(proto.GetInt64(c.StartTimestamp))/1000000000 ),
          "EndTimestamp": fmt.Sprintf("%f", float64(proto.GetInt64(c.EndTimestamp))/1000000000 ),
          "TimeNow": fmt.Sprintf("%d", time.Seconds()),
        }
  return smap
}

