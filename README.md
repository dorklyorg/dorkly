# Dorkly Flags Backend Bits
Not sure what's going on here? Check out the project overview [here](https://github.com/dorklyorg)

This repo contains the backend bits for the Dorkly Flags project:
1. Go code that runs in GitHub Actions converting human-friendly yaml files to a format that can be consumed by the ld-relay appliance.
2. Dockerfile + friends to build the image used in the deployed backend service.

### Docker
The Dockerfile is used to build the image used in the deployed backend service. It is built on top of the ld-relay image.
To build and publish it: (requires docker login with permissions to push to drichelson)
```bash
TAG=0.0.1 docker build --platform=linux/amd64 -t drichelson/dorkly:$TAG ./docker/ && docker push drichelson/dorkly:$TAG
```
