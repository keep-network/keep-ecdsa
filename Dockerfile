FROM golang:1.11.4-alpine3.7 AS runtime

ENV APP_NAME=keep-tecdsa \
	BIN_PATH=/usr/local/bin

FROM runtime AS gobuild

ENV GOPATH=/go \
	GOBIN=/go/bin \
	APP_NAME=keep-tecdsa \
	APP_DIR=/go/src/github.com/keep-network/keep-tecdsa \
	BIN_PATH=/usr/local/bin \
	# GO111MODULE required to support go modules
	GO111MODULE=on

RUN apk add --update --no-cache \
	g++ \
	git && \
	rm -rf /var/cache/apk/ && mkdir /var/cache/apk/ && \
	rm -rf /usr/share/man

RUN mkdir -p $APP_DIR

WORKDIR $APP_DIR

COPY ./ $APP_DIR/

RUN GOOS=linux go build -a -o $APP_NAME ./ && \
	mv $APP_NAME $BIN_PATH

FROM runtime

COPY --from=gobuild $BIN_PATH/$APP_NAME $BIN_PATH

# docker caches more when using CMD [] resulting in a faster build.
CMD []
