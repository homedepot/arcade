FROM alpine
RUN apk add --no-cache ca-certificates curl
COPY arcade /usr/local/bin
CMD ["/usr/local/bin/arcade"]
