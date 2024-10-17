# Home exam 2024

To build this project the go compiler is need which you can get here [Go](https://go.dev/)



All commands are assumed to be run in the repository directory

## Build and run 

```console
go build -o pointsalad ./src
./pointsalad -help
```

## Running server

```console
./pointsalad -server -bots 1 -players 1
```

## Running client

```console
./pointsalad -hostname localhost
```

## Test Point salad

```console
go test  ./src/pointsalad
```