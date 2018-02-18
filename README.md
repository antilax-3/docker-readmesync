# AntilaX-3/readme-sync
[![](https://images.microbadger.com/badges/version/antilax3/readme-sync.svg)](https://microbadger.com/images/antilax3/readme-sync "Get your own version badge on microbadger.com") [![](https://images.microbadger.com/badges/image/antilax3/readme-sync.svg)](https://microbadger.com/images/antilax3/readme-sync "Get your own image badge on microbadger.com") [![Docker Pulls](https://img.shields.io/docker/pulls/antilax3/readme-sync.svg)](https://hub.docker.com/r/antilax3/readme-sync/) [![Docker Stars](https://img.shields.io/docker/stars/antilax3/readme-sync.svg)](https://hub.docker.com/r/antilax3/readme-sync/)

[readme-sync](https://github.com/AntilaX-3/docker-readmesync) is a simple server that that provides an API to update a DockerHub repository's full description based on a specified GitHub repository's README.md, written in Node.js. 
## Usage
```
docker create --name=readmesync \
-v <path to config>:/config \
-p 80:80 \
antilax3/readme-sync
```
## Parameters
The parameters are split into two halves, separated by a colon, the left hand side representing the host and the right the container side. For example with a volume -v external:internal - what this shows is the volume mapping from internal to external of the container. So -v /mnt/app/config:/config would map /config from inside the container to be accessible from /mnt/app/config on the host's filesystem.

- `-v /config` - local path for readmesync config file
- `-p 80` - HTTP port for API webserver

It is based on alpine linux, utilising the official node docker repository with alpine tag, for shell access whilst the container is running do `docker exec -it readmesync /bin/bash`.

## Volumes

The container uses a single volume mounted at '/config'. This volume stores the configuration file 'readmesync.json'.

    config
    |-- readmesync.json

## Configuration

The readmesync.json is copied to the /config volume when first run. It has two mandatory parameters.

    dockerhub_username: String (Required) | Your DockerHub username
    dockerhub_password: String (Required) | Your DockerHub password

## Using the application

**API - GET command**

You can provide a GitHub branch if you want to sync a `README.md` from an branch other than master, if none if provided master is assumed.
```
http://<ip_address>:<port>/description/update?github_repo=<github_repo>&dockerhub_repo=<dockerhub_repo>
http://<ip_address>:<port>/description/update?github_repo=<github_repo>&github_branch=<github_branch>&dockerhub_repo=<dockerhub_repo>
```
## Version
- **18/02/18:** | Initial Release