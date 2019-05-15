FROM golang:1.12.5-alpine3.9 AS runtime

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

# Configure GitHub token to be able to get private repositories.
ARG GITHUBTOKEN
RUN git config --global url."https://$GITHUBTOKEN:@github.com/".insteadOf "https://github.com/"

# Get dependencies.
COPY go.mod $APP_DIR/
COPY go.sum $APP_DIR/

RUN go mod download

# Build the app.
COPY ./ $APP_DIR/

RUN GOOS=linux go build -a -o $APP_NAME ./ && \
	mv $APP_NAME $BIN_PATH

FROM runtime

COPY --from=gobuild $BIN_PATH/$APP_NAME $BIN_PATH

# docker caches more when using CMD [] resulting in a faster build.
CMD []
