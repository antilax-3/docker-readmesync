<p align="center">
  <a href="https://github.com/AntilaX-3/"><img src="https://avatars.githubusercontent.com/u/35715409" width="150" title="AntilaX-3"></a>
</p>

<p align="center">
  <a href="https://buildkite.com/antilax-3/readmesync"><picture><source media="(prefers-color-scheme: dark)" srcset="https://shieldcn.dev/badge/dynamic/json.svg?url=https%3A%2F%2Fimg.shields.io%2Fbuildkite%2F7a27cff7045d481072d214c0ad352e4d7b43c4a2de755341c9%2Fmaster.json&query=%24.message&label=build&logo=buildkite&logoColor=%2314cc80&mode=dark&size=sm&variant=outline"><img alt="Build" src="https://shieldcn.dev/badge/dynamic/json.svg?url=https%3A%2F%2Fimg.shields.io%2Fbuildkite%2F7a27cff7045d481072d214c0ad352e4d7b43c4a2de755341c9%2Fmaster.json&query=%24.message&label=build&logo=buildkite&logoColor=%2314cc80&mode=light&size=sm&variant=outline"></picture></a>
  <a href="https://www.gnu.org/licenses/lgpl-3.0"><picture><source media="(prefers-color-scheme: dark)" srcset="https://shieldcn.dev/github/antilax-3/docker-readmesync/license.svg?logo=gnu&logoColor=%23a42e2b&mode=dark&size=sm&variant=outline"><img alt="License" src="https://shieldcn.dev/github/antilax-3/docker-readmesync/license.svg?logo=gnu&logoColor=%23a42e2b&mode=light&size=sm&variant=outline"></picture></a>
  <a href="https://hub.docker.com/r/antilax3/readme-sync/tags"><picture><source media="(prefers-color-scheme: dark)" srcset="https://shieldcn.dev/badge/dynamic/json.svg?url=https%3A%2F%2Fimg.shields.io%2Fdocker%2Fimage-size%2Fantilax3%2Freadme-sync%2Flatest.json&query=%24.message&label=image%20size&logo=docker&logoColor=%232496ed&mode=dark&size=sm&variant=outline"><img alt="Docker Size" src="https://shieldcn.dev/badge/dynamic/json.svg?url=https%3A%2F%2Fimg.shields.io%2Fdocker%2Fimage-size%2Fantilax3%2Freadme-sync%2Flatest.json&query=%24.message&label=image%20size&logo=docker&logoColor=%232496ed&mode=light&size=sm&variant=outline"></picture></a>
  <a href="https://hub.docker.com/r/antilax3/readme-sync"><picture><source media="(prefers-color-scheme: dark)" srcset="https://shieldcn.dev/badge/dynamic/json.svg?url=https%3A%2F%2Fimg.shields.io%2Fdocker%2Fpulls%2Fantilax3%2Freadme-sync.json&query=%24.message&label=pulls&logo=docker&logoColor=%232496ed&mode=dark&size=sm&variant=outline"><img alt="Docker Pulls" src="https://shieldcn.dev/badge/dynamic/json.svg?url=https%3A%2F%2Fimg.shields.io%2Fdocker%2Fpulls%2Fantilax3%2Freadme-sync.json&query=%24.message&label=pulls&logo=docker&logoColor=%232496ed&mode=light&size=sm&variant=outline"></picture></a>
</p>

# AntilaX-3/readme-sync

[readme-sync](https://github.com/AntilaX-3/docker-readmesync) is a simple server that that provides an API to update a DockerHub repository's full description based on a specified GitHub repository's README.md, written in Go.
## Usage
```
docker create --name=readmesync \
-v <path to config>:/config \
-p 80:80 \
antilax3/readme-sync
```
## Tags

Two variants are built from the one Dockerfile, for `linux/amd64` and `linux/arm64`.

| Variant | Base | Tags |
| --- | --- | --- |
| wolfi | [antilax3/wolfi](https://hub.docker.com/r/antilax3/wolfi) | `latest` |
| alpine | [antilax3/alpine](https://hub.docker.com/r/antilax3/alpine) | `alpine` |

Wolfi is the default. Both variants run the same statically linked binary, so the choice between them is only the base. Every build is also tagged `BK<build>`, with `-alpine` appended for the alpine variant.

## Parameters
The parameters are split into two halves, separated by a colon, the left hand side representing the host and the right the container side. For example with a volume -v external:internal - what this shows is the volume mapping from internal to external of the container. So -v /mnt/app/config:/config would map /config from inside the container to be accessible from /mnt/app/config on the host's filesystem.

- `-v /config` - local path for readmesync config file
- `-p 80` - HTTP port for API webserver
- `-e PUID` - for UserID, see below for explanation
- `-e PGID` - for GroupID, see below for explanation
- `-e TZ` - for setting timezone information, eg Australia/Melbourne

It is based on wolfi, or alpine linux for the `alpine` tag, with s6 overlay, for shell access whilst the container is running do `docker exec -it readmesync /bin/bash`.

## User / Group Identifiers
Sometimes when using data volumes (-v flags) permissions issues can arise between the host OS and the container. We avoid this issue by allowing you to specify the user `PUID` and group `PGID`. Ensure the data volume directory on the host is owned by the same user you specify and it will "just work".

In this instance `PUID=1001` and `PGID=1001`. To find yours use `id user` as below:
`$ id <dockeruser>`
    `uid=1001(dockeruser) gid=1001(dockergroup) groups=1001(dockergroup)`
    
## Volumes

The container uses a single volume mounted at '/config'. This volume stores the configuration file 'readmesync.json'.

    config
    |-- readmesync.json

## Configuration

The readmesync.json is copied to the /config volume when first run, readable only by the container's user. It has two mandatory parameters and an optional port.

    dockerhub_username: String (Required) | Your DockerHub username
    dockerhub_password: String (Required) | Your DockerHub password, or a personal access token
    port:               Number (Optional) | The port the API listens on, 80 by default

Either the account password or a personal access token with read and write scope works. A token is recommended, as it can be scoped and revoked on its own, and it is required when the account has two-factor authentication or its organisation enforces SSO. Create one under Account settings, Personal access tokens.

## Using the application

**API - GET command**

You can provide a GitHub branch if you want to sync a `README.md` from a branch other than master, if none is provided master is assumed. Any path is accepted.
```
http://<ip_address>:<port>/description/update?github_repo=<github_repo>&dockerhub_repo=<dockerhub_repo>
http://<ip_address>:<port>/description/update?github_repo=<github_repo>&github_branch=<github_branch>&dockerhub_repo=<dockerhub_repo>
```

The response is `200 OK` once Docker Hub has stored the README, `400` for a missing or malformed field or a repository or branch with no `README.md`, and `502` when Docker Hub refuses the login or the update, with Docker Hub's own message in the body.
## Development

Linting runs locally through [lefthook](https://github.com/evilmartians/lefthook). Install the hooks once per clone:

```bash
lefthook install
```

`pre-commit` runs [golangci-lint](https://golangci-lint.run) and `go test`, [editorconfig-checker](https://github.com/editorconfig-checker/editorconfig-checker), [hadolint](https://github.com/hadolint/hadolint), `jq`, [shellcheck](https://github.com/koalaman/shellcheck), [typos](https://github.com/crate-ci/typos) and [yamllint](https://github.com/adrienverge/yamllint) over the staged files, and `commit-msg` enforces [Conventional Commits](https://www.conventionalcommits.org). Run everything on demand with:

```bash
lefthook run pre-commit --all-files
```

### Dependencies

readmesync uses the go standard library alone. Renovate bumps the `golang` build image in the Dockerfile and the go version in `go.mod`. The base images are followed at `antilax3/wolfi:latest` and `antilax3/alpine:latest`, so each build picks up their changes without a bump here.

## Version
- **30/09/26:** Rewrite readme-sync in Go and build it on the wolfi and alpine base images
- **30/09/26:** Build on wolfi by default and publish alpine under its own tag, for amd64 and arm64
- **04/07/25:** Updated to use alpine 3.22 image and s6 v3 service structure
- **22/02/18:** Updated to use alpine 3.7 image and build with jenkins
- **18/02/18:** Initial Release
