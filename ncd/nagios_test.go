package main

import (
	//  "strconv"
	"fmt"
	"testing"
	"os"
	"time"
	"bytes"
	proto "goprotobuf.googlecode.com/hg/proto"
)

var hostname = "test-host"
var servicename = "test-service"
var status = CheckStatus(CheckStatus_OK)

const co = "OK:\tcount to potato | potato=1\n hoooraaaay\n | potatoe=2"
const conp = "OK:\tcount to potato\n hoorraaaay\n "

var servicecheck = CheckResult{
	Hostname:       proto.String(hostname),
	ServiceName:    proto.String(servicename),
	Status:         &status,
	CheckOutput:    proto.String(co),
	StartTimestamp: proto.Int64(time.Nanoseconds() - 11111390),
	EndTimestamp:   proto.Int64(time.Nanoseconds()),
}
var hostcheck = CheckResult{
	Hostname:       proto.String(hostname),
	Status:         &status,
	CheckOutput:    proto.String(co),
	StartTimestamp: proto.Int64(time.Nanoseconds() - 11111390),
	EndTimestamp:   proto.Int64(time.Nanoseconds()),
}


type PerfElementCheck struct{ in, out string }
type PerfDataCheck struct{ in, outstr, outpd string }

var perfdatacheck = CheckResult{
	Hostname:       proto.String(hostname),
	ServiceName:    proto.String(servicename),
	Status:         &status,
	CheckOutput:    proto.String(conp),
	StartTimestamp: proto.Int64(time.Nanoseconds() - 11111390),
	EndTimestamp:   proto.Int64(time.Nanoseconds()),
	Perfdata: []*PerfData{
		&PerfData{Name: proto.String("potato"), Value: proto.Float32(1)},
		&PerfData{Name: proto.String("potatoe"), Value: proto.Float32(2)},
	},
}

var pdtests = [...]PerfDataCheck{
	{"OK: matched | time=2sec", "OK: matched ", "[name:\"time\" value:2 units:\"sec\" ]"},
	{"OK:\tcount to potato | potato=1\n hoooraaaay\n | potatoe=2", "OK:\tcount to potato \n hoooraaaay\n ", "[name:\"potato\" value:1  name:\"potatoe\" value:2 ]"},
	{"OK:\tcount to potato | potato=1\n hoooraaaay\n | potatoe=2%", "OK:\tcount to potato \n hoooraaaay\n ", "[name:\"potato\" value:1  name:\"potatoe\" value:2 units:\"%\" ]"},
}

func TestParsePerfDataElement(t *testing.T) {
	in := [...]PerfElementCheck{
		{"data=1elem;5;10;1;4", "name:\"data\" value:1 units:\"elem\" warning:5 critical:10 minimum:1 maximum:4 "},
		{"bananas=3", "name:\"bananas\" value:3 "},
	}
	for _, v := range in {
		pd, err := parsePerfDataElement(v.in)
		if err != nil {
			t.Error("parsePerfDataElement returned error on ", v, ": ", err)
		}
		out := proto.CompactTextString(pd)
		if v.out != out {
			t.Error("mismatch for", v.in, ":\n", out, "\n", v.out)
		}
	}
}

func TestParsePerfData(t *testing.T) {
	for i, v := range pdtests {
		outstr, outpd, err := ParseRawPluginOutput(v.in)
		if err != nil {
			t.Errorf("test %d: can't parse plugin output: %v", i, err)
		}
		if outstr != v.outstr {
			t.Errorf("test %d: checkoutput doesn't match:\noutput:   %#v\n --- \nexpected: %#v", i, outstr, v.outstr)
		}
		t.Log("outpd:", outpd)
		outpdstr := fmt.Sprint(outpd)
		if v.outpd != outpdstr {
			t.Errorf("test %d: perfdata doesn't match:\noutput:   %#v\n --- \nexpected: %#v", i, outpdstr, v.outpd)
			t.Logf("%#v\n", outpdstr)
		}
	}
}

func TestGetIntFromEnum(t *testing.T) {
	blarg := NewCheckStatus(CheckStatus_WARNING)
	t.Logf("%#v\n", blarg)
	t.Logf("%#v\n", *blarg)
	t.Logf("%s\n", *blarg)
}

func TestStringMap(t *testing.T) {
	/* fmt.Printf("check = %#v\n", servicecheck)*/
}

func TestStringEscaping(t *testing.T) {
	buf := new(bytes.Buffer)
	co := "OK:\tcount to potato | potato=1\n hoooraaaay\n | potatoe=2"
	escapedStringFormatter(buf, "", co)
	for i := 0; i < buf.Len(); i++ {
		c, err := buf.ReadByte()
		if err != nil && err == os.EOF {
			break
		}
		switch c {
		case '\n', '\t':
			t.Error("failed escaping")
		}
	}
}

func TestCreateSpoolFile(t *testing.T) {
	f, err := spoolFile(".", "test")
	if err != nil {
		t.Error("spoolFile:", err)
	}
	fi, err := f.Stat()
	if err != nil {
		t.Error("spoolFile: couldn't stat file:", err)
	}
	_ = fi
	err = os.Remove(f.Name())
	if err != nil {
		err := os.Remove(f.Name())
		if err != nil {
			t.Log("couldn't remove", f.Name(), ":", err)
		}
	}
}

func TestWriteServiceCheckFile(t *testing.T) {
	_test_writeCheckFile(t, &servicecheck)
}

func TestWriteHostFile(t *testing.T) {
	_test_writeCheckFile(t, &hostcheck)
}

func _test_writeCheckFile(t *testing.T, check *CheckResult) {
	fn, err := WriteCheck(check, ".")
	t.Log("wrote check to", fn)
	if err != nil {
		t.Error("WriteCheck:", err)
	}
	for _, v := range [...]string{fn, fn + ".ok"} {
		err := os.Remove(v)
		if err != nil {
			err := os.Remove(v)
			if err != nil {
				t.Log("couldn't remove", v, ":", err)
			}
		}
	}
	return
}

func BenchmarkStringEscaping(b *testing.B) {
	for i := 0; i < b.N; i++ {
		fmt.Printf("hello world\n")
	}
}
