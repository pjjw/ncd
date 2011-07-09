# A sample top level Makefile for multi package Go projects.

include $(GOROOT)/src/Make.inc

CMDS=\
        send_check\
        ncd

all: make

make: $(addsuffix .make, $(CMDS))
clean: $(addsuffix .clean, $(CMDS))

%.install:
	$(MAKE) -C $* install

# compile all packages before any command
%.make:
	$(MAKE) -C $*

# establish dependancies between packages
#package-2.install: package-1.install
#package-1.install package-2.install: package-3.install

%.clean:
	$(MAKE) -C $* clean

%.nuke:
	$(MAKE) -C $* nuke
