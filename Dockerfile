FROM golang:1.13.8-alpine3.10 AS gobuild

ARG VERSION
ARG REVISION

ENV GOPATH=/go \
	GOBIN=/go/bin \
	APP_NAME=keep-ecdsa \
	APP_DIR=/go/src/github.com/keep-network/keep-ecdsa \
	BIN_PATH=/usr/local/bin \
	# GO111MODULE required to support go modules
	GO111MODULE=on

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
RUN cd /go/pkg/mod/github.com/gogo/protobuf@v1.3.1/protoc-gen-gogoslick && go install .
RUN cd /go/pkg/mod/github.com/ethereum/go-ethereum@v1.9.10/cmd/abigen && go install .

# Install Solidity contracts.
COPY ./solidity $APP_DIR/solidity
RUN cd $APP_DIR/solidity && npm install

# Generate code.
COPY ./pkg/chain/gen $APP_DIR/pkg/chain/gen
COPY ./pkg/ecdsa/tss/gen $APP_DIR/pkg/ecdsa/tss/gen
# Need this to resolve imports in generated Ethereum commands.
COPY ./config $APP_DIR/config
RUN go generate ./.../gen

# Build the application.
COPY ./ $APP_DIR/

# Configure private repositories for Go dependencies
ARG GOPRIVATE

RUN GOOS=linux GOPRIVATE=$GOPRIVATE go build -ldflags "-X main.version=$VERSION -X main.revision=$REVISION" -a -o $APP_NAME ./ && \
	mv $APP_NAME $BIN_PATH

# Configure runtime container.
FROM alpine:3.10

ENV APP_NAME=keep-ecdsa \
	BIN_PATH=/usr/local/bin

COPY --from=gobuild $BIN_PATH/$APP_NAME $BIN_PATH

# docker caches more when using CMD [] resulting in a faster build.
CMD []
