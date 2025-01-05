# syntax=docker/dockerfile:1

FROM debian:bookworm-slim AS builder
SHELL ["/bin/bash", "-c"]
ENV MISE_VERSION=v2024.9.10

RUN apt-get update && \
    apt-get upgrade -y && \
    apt-get install -y curl && \
    curl https://mise.run | sh && \
    mv /root/.local/bin/mise /usr/sbin/mise

COPY ./.mise.toml .

RUN mise trust && mise install --yes && \
    eval "$(mise activate bash)"

COPY go.mod go.sum ./

RUN eval "$(mise hook-env --shell=bash)" && \
    go mod download && go mod verify

COPY . .

RUN eval "$(mise hook-env --shell=bash)" && \
    go build -o management-service ./cmd/management-service/main.go

FROM debian:bookworm-slim

ENV USER_ID=65535
ENV GROUP_ID=65535
ENV USER_NAME=ponix
ENV GROUP_NAME=ponix

RUN apt-get update && apt-get upgrade -y && apt-get install -y ca-certificates

COPY --from=builder /management-service /home/ponix/management-service
COPY --from=builder /root/.local/share/mise/installs/atlas/latest/bin/atlas /home/ponix/atlas

RUN useradd -ms /bin/bash ponix

USER ponix
WORKDIR /home/ponix

ENV PATH="/home/ponix"

ENTRYPOINT ["./management-service"]
