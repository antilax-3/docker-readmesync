# syntax=docker/dockerfile:1
ARG BASE_IMAGE="antilax3/node:latest"

# The bundle and every package it requires are plain javascript, with no native addon among them, so they are built
# once on the build platform and copied into the image of each target platform unchanged.
FROM --platform=${BUILDPLATFORM} ${BASE_IMAGE} AS build

WORKDIR /app

COPY root/app/ ./

SHELL ["/bin/ash", "-euo", "pipefail", "-c"]

RUN <<'EOF'
set -euo pipefail

echo "**** build node application ****"
npm install
npm run build

echo "**** keep only the runtime dependencies ****"
# backpack bundles src/ into build/main.js and leaves every package it requires external, so the image needs those
# packages and none of the toolchain that built the bundle.
npm prune --omit=dev
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
COPY --link root/etc/ /etc/
COPY --link --from=build /app/build/main.js /app/main.js
COPY --link --from=build /app/node_modules/ /app/node_modules/

# ports and volumes
EXPOSE 80
VOLUME /config
