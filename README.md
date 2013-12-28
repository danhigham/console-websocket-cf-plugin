### console-websocket-cf-plugin

This is a [CF](https://github.com/cloudfoundry/cf) plugin for connecting to the STDIN / STDOUT of a remote process (rails console for example) on Cloud Foundry whilst simultaneously running the intended application in the same container.
It uses a binary written in Go on the server side to provide the websocket connection and also serve the application itself, if you wish to serve that up too.

## TL;DR Rails example

1. Install the Gem
