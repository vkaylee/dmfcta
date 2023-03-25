FROM debian:buster
RUN apt-get update && \
    apt-get install -y entr && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*
RUN mkdir /app
WORKDIR /app
COPY do.sh /entrypoint
ENTRYPOINT [ "/entrypoint" ]