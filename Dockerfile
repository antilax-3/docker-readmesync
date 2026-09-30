# syntax=docker/dockerfile:1
ARG BASE_IMAGE="antilax3/wolfi:latest"

# readmesync is a static go binary, so it is cross-compiled once per target on the build platform rather than built
# under emulation, and the same binary runs on either base.
FROM --platform=${BUILDPLATFORM} golang:1.27-alpine AS build

ARG TARGETOS
ARG TARGETARCH

WORKDIR /src

COPY go.mod ./
COPY cmd/ ./cmd/
COPY internal/ ./internal/

SHELL ["/bin/ash", "-euo", "pipefail", "-c"]

RUN <<'EOF'
set -euo pipefail

echo "**** install build packages ****"
apk add --no-cache libcap-setcap

echo "**** test readmesync ****"
go test ./...

echo "**** build readmesync ****"
CGO_ENABLED=0 GOOS="${TARGETOS}" GOARCH="${TARGETARCH}" go build -trimpath -buildvcs=false -ldflags="-s -w" \
  -o /out/app/readmesync ./cmd/readmesync

# The service runs as abc and listens on port 80 by default, which only a process allowed to bind privileged ports
# can do.
setcap cap_net_bind_service=+ep /out/app/readmesync
EOF

FROM ${BASE_IMAGE}

# set version labels
ARG build_date
ARG version
LABEL build_date="${build_date}"
LABEL version="${version}"
LABEL maintainer="Nightah"

# set working directory
WORKDIR /app

# copy local files
COPY --link root/ /
COPY --link --from=build /out/ /

# ports and volumes
EXPOSE 80
VOLUME /config
