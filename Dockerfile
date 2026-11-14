# One image for both controller and worker, so each runs as its own
# "node" with its own network interface (needed for tc/netem per node).

# Stage 1: build both Go binaries
FROM golang:1.24-alpine AS build
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -o /out/controller ./controller && \
    CGO_ENABLED=0 go build -o /out/worker ./worker

# Stage 2: runtime
#   docker CLI  -> worker runs jobs as containers on the host's Docker
#   iproute2    -> tc/netem (delay, loss)
#   iptables    -> drop incoming packets during an outage
FROM docker:27-cli
RUN apk add --no-cache iproute2 iptables
COPY --from=build /out/controller /out/worker /usr/local/bin/
ENTRYPOINT []
