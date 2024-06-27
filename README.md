![go workflow](https://github.com/dorklyorg/dorkly/actions/workflows/go.yml/badge.svg)
# Dorkly Flags Backend Bits
[More info on the Dorkly project](https://github.com/dorklyorg/dorkly/wiki)

This repo contains the backend bits for the Dorkly Flags project:
1. Go code that runs in GitHub Actions converting human-friendly yaml files to a format that can be consumed by the ld-relay appliance.
2. Dockerfile + friends to build the image used in the deployed backend service.

### Docker
The Dockerfile is used to build the image used in the deployed backend service. It is built on top of the ld-relay image.
To build and publish it: (requires docker login with permissions to push to drichelson)
```bash
export TAG=0.0.6 && docker build --platform=linux/amd64 -t drichelson/dorkly:$TAG ./docker/ && docker push drichelson/dorkly:$TAG
```
