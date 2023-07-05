FROM debian:buster
RUN apt-get update && \
    apt-get install -y entr && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*
RUN mkdir /app /applogs
WORKDIR /app
COPY do.sh /entrypoint
ENTRYPOINT [ "/entrypoint" ]


FROM golang:alpine AS dev
WORKDIR /app
RUN apk add npm
RUN npm install -g nodemon

FROM golang:alpine AS build
# Set destination for COPY
WORKDIR /app
# Download Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code. Note the slash at the end, as explained in
# https://docs.docker.com/engine/reference/builder/#copy
COPY *.go ./
COPY action/*.go ./action/
COPY apputils/*.go ./apputils/
# Build
RUN GOOS=linux CGO_ENABLED=0 go build -ldflags="-extldflags=-static" -o /myapp


FROM debian:buster AS app
RUN apt-get update && \
    apt-get install -y entr && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /appspace
COPY --from=build /myapp ./app
ENTRYPOINT [ "./app" ]