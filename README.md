# taikai

Taikai is open source community and events software.

Taikai is a japanese word that can mean Large Meeting or Convention - 大会

## History

Forge Utah Foundation is a local tech community in Utah built for the developers, engineers, data scientists-- for the true builders of tech. We host several free and open developer groups and we are a big user of Meetup.com. Unfortunately, Meetup doesn't provide discounts for not-for-profit organizations and the cost has become extremely prohibitive. Since we are a group of technologists, we thought "let's build our own!". Hopefully other communities will find it beneficial as well.

## Contributing

You can use the [devcontainer](https://containers.dev/) config to get going with all the needed dependencies. But here is the stack:

Generated via copier from github.com/catalystsquad/copier-go-cobra-app

In order to run this code you will need the following on your path:

* buf
* Go 1.21+
* Node 18+ (and npm of course)
* Java of some kind (for the OpenAPI Client Generator)
* bash, which you should just have anyway

## Pull Requests

We have three checks that must pass before a PR can be merged:

* `go test ./...`
* `buf check lint`
* `golangci-lint run`
* GitGuardian

To run these checks locally, you can use the `./tools.sh`.

# Running

If using protos and the protos have not been built or are not up to date, you will be sad. `./tools.sh build_protos` will handle that.

If you use skaffold, make sure you have the helm repos added (look at skaffold.yaml for the command to run to do so).

Otherwise, you can `./tools.sh run` or `skaffold dev` if you prefer. Running via tools will run it with `go` and skaffold will run in a kubernetes cluster.
