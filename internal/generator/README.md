# generator

The generator takes numerous sources of a Lagoon build (files sourced from the git checkout, build pod environment variables, lagoon variables, etc.) and pre-calculates the build into a set of values.

These values are then able to be passed to various templating tools and used to create the kubernetes (or other) resources required to create a build.

The basic idea is that the calculated values are final, before any templating takes place. This is to help ensure that builds are validated before any work is under taken.

Some components of the generator are as follows

* backup and restore
* docker-compose to lagoon service
* ingress