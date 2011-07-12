package main

import (
	"encoding/base64"
	"goprotobuf.googlecode.com/hg/proto"
	"log"
	"testing"
	"json"
)

type jsonTest struct {
	cr   *CheckResult
	json string
}

var jsonTests = []jsonTest{
	{&CheckResult{Hostname: proto.String("testhost.foo.bar"),
		ServiceName:    proto.String("test-service"),
		Status:         NewCheckStatus(CheckStatus_OK),
		StartTimestamp: proto.Int64(151892832),
		EndTimestamp:   proto.Int64(151895000)},
		"{\"Hostname\":\"testhost.foo.bar\",\"ServiceName\":\"test-service\",\"Status\":0,\"CheckPassive\":null,\"CheckScheduled\":null,\"CheckOutput\":null,\"StartTimestamp\":151892832,\"EndTimestamp\":151895000,\"XXX_unrecognized\":\"\"}"}}

var jsonStr = "{\"Hostname\":\"testhost.foo.bar\",\"ServiceName\":\"test-service\",\"Status\":0,\"CheckPassive\":null,\"CheckScheduled\":null,\"CheckOutput\":null,\"StartTimestamp\":151892832,\"EndTimestamp\":151895000,\"XXX_unrecognized\":\"\"}"

var cr = CheckResult{
	Hostname:       proto.String("testhost.foo.bar"),
	ServiceName:    proto.String("test-service"),
	Status:         NewCheckStatus(CheckStatus_OK),
	StartTimestamp: proto.Int64(151892832),
	EndTimestamp:   proto.Int64(151895000),
}

var crs = CheckResultSet{
	Results: []*CheckResult{&cr},
}

func TestMarshalUnmarshal(t *testing.T) {
	t.Log("marshalling protobuf")
	buf, err := proto.Marshal(&cr)
	if err != nil {
		log.Fatal("marshal error: ", err)
	}
	t.Log("marshalled")
	rcr := new(CheckResult)
	t.Log("unmarshalling")
	err = proto.Unmarshal(buf, rcr)
	t.Log(rcr)
}

func TestJsonToProto(t *testing.T) {
	for i, v := range jsonTests {
		ncr := &CheckResult{}
		jsonBytes := []byte(v.json)
		t.Logf("test %d: unmarshalling json to protobuf struct", i)
		err := json.Unmarshal(jsonBytes, ncr)
		if err != nil {
			t.Errorf("test %d: %s", i, err)
		}
		if proto.CompactTextString(ncr) != proto.CompactTextString(v.cr) {
			t.Errorf("test %d: mismatch\noutput:    %s\nexpected: %s", i, ncr, v.cr)
		}
		t.Logf("%s\n", proto.CompactTextString(ncr))
	}
}

func TestJsonOutput(t *testing.T) {
	t.Log("marshalling protobuf as json")
	buf, err := json.Marshal(&crs)
	t.Logf("%#v\n", string(buf))
	t.Log(err)
}

func TestMarshalUnmarshalBase64(t *testing.T) {
	var encbuf []byte
	var decbuf []byte
	t.Logf("start with buf %s\n", proto.CompactTextString(&cr))
	t.Log("marshalling protobuf")
	buf, err := proto.Marshal(&cr)
	if err != nil {
		t.Error("marshal error: ", err)
	}
	t.Log("marshalled")
	t.Log("urlencoding")
	t.Logf("need %d size buffer\n", base64.URLEncoding.EncodedLen(len(buf)-1))
	t.Log(buf)
	t.Logf("%v %s\n", buf, buf)
	encbuf = make([]byte, base64.URLEncoding.EncodedLen(len(buf)), base64.URLEncoding.EncodedLen(len(buf)))
	base64.URLEncoding.Encode(encbuf, buf)
	t.Log("urlencoded")
	t.Log("urldecoding")
	t.Logf("need %d size buffer\n", base64.URLEncoding.DecodedLen(len(encbuf)))
	t.Logf("%v %s\n", encbuf, encbuf)
	decbuf = make([]byte, base64.URLEncoding.DecodedLen(len(encbuf)), base64.URLEncoding.DecodedLen(len(encbuf)))
	n, err := base64.URLEncoding.Decode(decbuf, encbuf)
	t.Logf("wrote %d bytes from encbuf to decbuf. len(encbuf)=%d, len(buf)=%d\n", n, len(encbuf), len(buf))
	if err != nil {
		t.Error("urldecode error: ", err)
	}
	t.Log("urldecoded")

	t.Log(buf, decbuf)

	rcr := &CheckResult{}
	t.Log("unmarshalling")
	err = proto.Unmarshal(decbuf, rcr)
	t.Logf("%s\n", proto.CompactTextString(rcr))
}
