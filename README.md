# Copy Trade app

## Ubuntu 18.04 

This is a golang structure was cloned from https://github.com/golang-standards/project-layout. 

# DEV environment.
* MacOS Catalina 10.15.1
* Clang version 11.0.0
* Go 1.14.6
* Docker version 19.03.2, build 6a30dfc
* docker-compose version 1.24.1, build 4667896b


## Technologies and Libraries.

* Cobra [`Cobra`](https://github.com/spf13/cobra)
* Logrus [`Logrus`](https://github.com/sirupsen/logrus)
* Moby [`Go moby`] project (https://github.com/moby/moby)


## How to build locally
```bash
go build github.com/vietnamz/prime-generator/cmd/hello
```
## How to build from local.

```bash
./scripts/build.sh
```

## How to build dependencies on clean ubuntu 18.04 linux machine.

```bash
./scripts/build-dependencies.sh
```
## How to run.
```bash
./bin/hello
```

## How to use docker
at the root folder run below command. docker and docker compose are required.
```bash
docker-compose up -d --build
```
