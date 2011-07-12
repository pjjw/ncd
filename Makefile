include $(GOROOT)/src/Make.inc

TARG=ncd

GOFILES=\
				ncd.go\
				checkresult.pb.go\
				spoolfile.go\
				templates.go\
				send_check.go\
				nagios.go

include $(GOROOT)/src/Make.cmd
include $(GOROOT)/src/pkg/goprotobuf.googlecode.com/hg/Make.protobuf
