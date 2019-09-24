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
	make \
	nodejs \
	npm \
	python \
	git \ 
	protobuf && \
	rm -rf /var/cache/apk/ && mkdir /var/cache/apk/ && \
	rm -rf /usr/share/man

# Install Solidity compiler.
COPY --from=ethereum/solc:0.5.8 /usr/bin/solc /usr/bin/solc

# Configure GitHub token to be able to get private repositories.
ARG GITHUBTOKEN
RUN git config --global url."https://$GITHUBTOKEN:@github.com/".insteadOf "https://github.com/"

# Configure working directory.
RUN mkdir -p $APP_DIR
WORKDIR $APP_DIR

# Get dependencies.
COPY go.mod $APP_DIR/
COPY go.sum $APP_DIR/

RUN go mod download

# Install code generators.
RUN cd /go/pkg/mod/github.com/keep-network/go-ethereum@v1.8.27/cmd/abigen && go install .
RUN cd /go/pkg/mod/github.com/gogo/protobuf@v1.2.1/protoc-gen-gogoslick && go install .

# Install Solidity contracts.
COPY ./solidity $APP_DIR/solidity
RUN cd $APP_DIR/solidity && npm install

# Generate code.
COPY ./pkg/chain/eth/gen $APP_DIR/pkg/chain/eth/gen
COPY ./pkg/registry/gen $APP_DIR/pkg/registry/gen
RUN go generate ./.../gen

# Build the application.
COPY ./ $APP_DIR/

RUN GOOS=linux go build -a -o $APP_NAME ./ && \
	mv $APP_NAME $BIN_PATH

# Configure runtime container.
FROM runtime

COPY --from=gobuild $BIN_PATH/$APP_NAME $BIN_PATH

# docker caches more when using CMD [] resulting in a faster build.
CMD []
