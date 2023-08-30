.PHONY: all build clean run check cover lint docker help

dateTime=`date +%F_%T`
ARCH="linux-amd64"

all: build

build:
	mkdir -p build
	xgo -targets=linux/amd64 -ldflags="-w -s" -out=build/telegram-premium -pkg=cmd/telegram-premium/main.go .
	tar czvf build/telegram-premium_${dateTime}.tar.gz \
		build/telegram-premium-${ARCH} \
		template \
		assets \
		restart.sh
