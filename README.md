### console-websocket-cf-plugin

This is a [CF](https://github.com/cloudfoundry/cf) plugin for connecting to the STDIN / STDOUT of a remote process (rails console for example) on Cloud Foundry whilst simultaneously running the intended application in the same container.
It uses a binary written in Go on the server side to provide the websocket connection and also serve the application itself, if you wish to serve that up too.

## TL;DR example

1 - Install the console-websocket-cf-plugin gem

```
$ gem install console-websocket-cf-plugin
```

2 - Copy the pre-built linux binary from the Github repo in to the root of your Rails app

```
$ cd my_rails_app
$ wget https://github.com/danhigham/console-websocket-cf-plugin/blob/master/console-server/console-server-linux-amd64?raw=true
```

3 - Modify the application manifest, note the 'command' property

```yml
---
applications:
- name: rails-console-test
  memory: 256M
  instances: 1
  host: rails-console-test
  domain: cfapps.io
  path: .
  command: ./console-server-linux-amd64 -console-process="rails c" -main-process="bundle exec rails s -p 8080"
```

The processes that are started by the console-server binary are completely configurable, you could for example just run 'bash' for the console process. If you wish to make the application available, mounted at /, then it needs to be bound to port 8080, this is the port console-server expects to proxy requests to.

4 - Push the app

```
$ cf push --reset
```

5 - Start a console
```
$ cf console <app name>
```