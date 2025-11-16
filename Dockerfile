FROM golang:1.25.1

WORKDIR /usr/src/server
COPY . .

RUN make build
CMD ["./server"]
