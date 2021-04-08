FROM golang:1.13.8-alpine3.10 AS gobuild

# HOST_CHAIN argument defines the chain implementation which should be used
# during image build process.
ARG HOST_CHAIN=ethereum

# Several host chain Go modules which use native C code underneath may
# need to know the C standard library implementation used by the platform.
# The LIBC env variable specifies that information and is used to pass it
# to the Go compiler via build tags. The default value of LIBC variable
# is `musl` as it is the C standard library implementation used on Alpine.
ENV LIBC=musl

ENV GOPATH=/go \
	GOBIN=/go/bin \
	APP_NAME=keep-ecdsa \
	APP_DIR=/go/src/github.com/keep-network/keep-ecdsa \
	BIN_PATH=/usr/local/bin \
	# GO111MODULE required to support go modules
	GO111MODULE=on \
	APP_BUILD_TAGS="$HOST_CHAIN $LIBC" \
	ABIGEN_BUILD_TAGS=$LIBC

RUN apk add --update --no-cache \
	g++ \
	linux-headers \
	make \
	nodejs \
	npm \
	python \
	git \
	protobuf && \
	rm -rf /var/cache/apk/ && mkdir /var/cache/apk/ && \
	rm -rf /usr/share/man

# Install Solidity compiler.
COPY --from=ethereum/solc:0.5.17 /usr/bin/solc /usr/bin/solc

# Get gotestsum tool
RUN go get gotest.tools/gotestsum

# Configure GitHub token to be able to get private repositories.
ARG GITHUB_TOKEN
RUN git config --global url."https://$GITHUB_TOKEN:@github.com/".insteadOf "https://github.com/"

# Configure working directory.
RUN mkdir -p $APP_DIR
WORKDIR $APP_DIR

# Get dependencies.
COPY go.mod $APP_DIR/
COPY go.sum $APP_DIR/

RUN go mod download

# Install code generators.
RUN cd /go/pkg/mod/github.com/gogo/protobuf@v1.3.2/protoc-gen-gogoslick && go install .

# Install Solidity contracts.
COPY ./solidity $APP_DIR/solidity
RUN cd $APP_DIR/solidity && npm install

# Generate code.
COPY ./pkg/chain/gen/$HOST_CHAIN $APP_DIR/pkg/chain/gen/$HOST_CHAIN
COPY ./pkg/ecdsa/tss/gen $APP_DIR/pkg/ecdsa/tss/gen
# Need this to resolve imports in generated chain commands.
COPY ./config $APP_DIR/config
RUN go generate ./...

# Build the application.
COPY ./ $APP_DIR/

# Cleanup the `pkg/chain/gen` dir from unused chains bindings. Leave only
# the ones which are currently in use. This helps reducing the size of
# resulting binary and can prevent unexpected errors which may occur due to
# transitive dependencies conflicts.
RUN cd $APP_DIR/pkg/chain && \
	mv ./gen/$HOST_CHAIN ./temp && \
	rm -rf ./gen && \
	mkdir ./gen && \
	mv ./temp ./gen/$HOST_CHAIN

# Client Versioning.
ARG VERSION
ARG REVISION

RUN GOOS=linux go build -tags "$APP_BUILD_TAGS" -ldflags "-X main.version=$VERSION -X main.revision=$REVISION" -a -o $APP_NAME ./ && \
	mv $APP_NAME $BIN_PATH

# Configure runtime container.
FROM alpine:3.10

ENV APP_NAME=keep-ecdsa \
	BIN_PATH=/usr/local/bin

COPY --from=gobuild $BIN_PATH/$APP_NAME $BIN_PATH

# docker caches more when using CMD [] resulting in a faster build.
CMD []
