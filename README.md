# keep-tecdsa

## Build

To build execute a command:
```sh
go build .
```

## Test

To test execute a command:
```sh
go test ./...
```

## Docker

To build a Docker image execute a command:
```sh
docker build -t keep-tecdsa .
```

To run execute a command:
```sh
docker run -it keep-tecdsa keep-tecdsa sign <hash>
```
Where `<hash>` is a message to sign.

## Run

To run execute a command:
```sh
./keep-tecdsa sign <hash>
```
With `<hash>` as a string to sign.

Sample output:
```sh
➜  keep-tecdsa git:(signer) ./keep-tecdsa sign YOLO
--- Generated Public Key:
X: 2295845dbe5b5af2b516afa990e9113073793b6f861b66aa36e453e3a0e976f1
Y: d6d1923fa28c29d9fc2eb274cb54efc16875fab6d2d741e56a8afc7783e3f03b
--- Signature:
R: 6479fff99d7aa3f22d9b489f164a6e904abdb74d6cc44fa5b274903accee366a
S: 3ae8dcb534aa12c84214e7f448c5f60dbc048c64e60977d2b0a81b76cece76c8
```
