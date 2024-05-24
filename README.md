# dorkly
Current status: not ready for public consumption. Stay tuned.

# Project Description
Open Source Feature Flag system.
Dorkly is a git-based open-source feature flag backend for [LaunchDarkly](https://launchdarkly.com/features/feature-flags/) SDKs.
This project implements a small but valuable subset of LaunchDarkly's featureset. It is not a drop-in replacement for LaunchDarkly. 
It is intended to be a simple, open-source alternative for small projects that don't need all of LaunchDarkly's features or are unable to use LaunchDarkly's backend for compliance reasons.
It consists of:
1. A GitHub Action) that reads in human-friendly yaml files, and converts them to an archive format consumed by:
2. A Docker image that runs a backend service that serves the flags to your application. This backend service is a [very thin wrapper](docker/Dockerfile) around the [ld-relay](https://docs.launchdarkly.com/sdk/relay-proxy) appliance running in [offline mode](https://docs.launchdarkly.com/sdk/relay-proxy/offline)
3. A [Terraform module](https://registry.terraform.io/modules/dorklyorg/dorkly-flags/aws/latest) that provisions the backend service on AWS and sets up the necessary permissions, and a GitHub repository to store the flags.

# Getting Started: One time setup
## First steps
1. Determine your project scope and come up with a short name. [Helpful doc](https://docs.launchdarkly.com/home/getting-started/vocabulary#project)
2. Determine your starting environments. These can be changed later so it's ok to use the defaults. [Helpful doc](https://docs.launchdarkly.com/home/getting-started/vocabulary#environment)
3. Provision your infrastructure using the [Terraform module](https://registry.terraform.io/modules/dorklyorg/dorkly-flags/aws/latest). [Example](https://github.com/dorklyorg/terraform-aws-dorkly-flags/blob/main/examples/main/main.tf)

## Setting up your application with a properly configured LaunchDarkly SDK: Server-side (golang example)
TODO: terraform: autogenerate example snippets including the sdk key and backend service url instead of this manual process.
1. Set up your application with the LaunchDarkly SDK. [Helpful doc](https://docs.launchdarkly.com/sdk/server-side)
2. For a quick example check out the [hello-go example](https://github.com/launchdarkly/hello-go/blob/main/main.go#L35) program.
3. *Not yet implemented but needed for MVP*: Grab the sdk key from either AWS secrets manager or if it is a non-production environment, from the GitHub repo. You'll use it in the next step.
4. Using the hello-go example as a starting point, set the `LAUNCHDARKLY_SDK_KEY` environment variable to the SDK key you used when provisioning the infrastructure.
5. *Not yet implemented but needed for MVP*: Grab the backend service url from the GitHub repo's readme/other file tbd. You'll use it in the next step.
6. Instead of initializing a default client, initialize a client with the url of the backend service:
```golang
	dorklyConfig := ld.Config{
		ServiceEndpoints: ldcomponents.RelayProxyEndpoints("<YOUR_DORKLY_URL>"),
		Events:           ldcomponents.NoEvents(),
	}

	ldClient, err := ld.MakeCustomClient(dorklySdkKey, dorklyConfig, 5*time.Second)
```

# Common Tasks
## Adding a feature flag
In your newly created GitHub repo you'll notice some example yaml files under the `project/` directory. These are intended to be a starting point, so you can create additional yaml files for your own flags.
1. Create a flag overview file for the project. Each flag's basic properties are defined at the project level. This includes name, description, and the type of flag (boolean, booleanRollout, etc). This file should be created in the `project/flags` directory. The naming convention is `<flagName>.yml`. Check out the example flags as a starting point.
2. Create environment-specific flag config files. Under each environment directory in `project/environments`, create a file with the same name as the flag file. This file will contain the environment-specific configuration for the flag. The naming convention is `<flagName>.yml`. Check out the example flags as a starting point.
3. *Not yet implemented but needed for MVP*: Pull Request checks will validate your changes for well-formedness.
4. Commit your changes to the main branch. The GitHub Action will automatically pick up the changes and update the backend service.

## Changing a feature flag for an environment
1. Update the flag file in the environment directory.
2. Commit your changes to the main branch. The GitHub Action will automatically pick up the changes and update the backend service.

## Adding an environment
1. Navigate to your Terraform config
2. Add a new environment to the `environments` variable and execute the changes.
3. Once your Terraform run has been applied, you can add flag configs for the environment manually, or by copying the contents of an existing environment.

### Process:
1. Load config
2. Load local yaml files
3. Convert to ld structs
4. Fetch/load existing relay archive: either local files (testing), or from S3 (production)
5. Reconcile local files with relay archive. Update versions etc as needed.
6. Marshal json files
7. Create checksum of json files
8. Create ld-relay archive of json + checksum files
9. Save locally (testing) or upload to S3 (production)


### Go code
The go code in this repo runs in GitHub Action provisioned by the [dorklyorg/dorkly-flags Terraform module](https://registry.terraform.io/modules/dorklyorg/dorkly-flags/aws/latest).

### Docker
The Dockerfile is used to build the image used in the deployed backend service. It is built on top of the ld-relay image.
To build and publish it: (requires docker login with permissions to push to drichelson)
```bash
TAG=0.0.1 docker build --platform=linux/amd64 -t drichelson/dorkly:$TAG ./docker/ && docker push drichelson/dorkly:$TAG
```

### Ideas for later:
1. Create GitHub releases with human-readable diffs of flag changes
2. https://full-stack.blend.com/how-we-write-github-actions-in-go.html#introduction
2. Part of GitHub actions: connect to relay and await changes.
3. PR check to ensure well-formed yaml files
4. Publish a GitHub action? then we don't need to check in a binary
5. Send Slack notifications on changes.


### Current functionality is limited to a subset of LaunchDarkly's feature flag types.
Here's what is supported:
- One project per git repo. If you need more projects create more repos.
- Boolean flags: either on or off, or a percent rollout based on user id
- Only the user context kind is supported
- server-side flags and client-side flags (can exclude client-side on a per-flag basis)

Ideas for later later
1. Segments
2. String variation flags
3. Number variation flags
4. Json variation flags
