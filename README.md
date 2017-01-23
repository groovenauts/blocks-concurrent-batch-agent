# blocks-concurrent-batch-agent

## Overview

`blocks-concurrent-batch-agent` is an agent for `concurrent bacth board` of [magellan-blocks](https://www.magellanic-clouds.com/blocks/).
`blocks-concurrent-batch-agent` builds your managed instance groups and
watches metrics of your pubsub subscriptions. When the number of the unsubscribed messages
increases or descreases, `blocks-concurrent-batch-agent` detects it and scale your managed instance group.

## Setup

1. Install Go
  - https://golang.org/doc/install
  - Or use [goenv](https://github.com/kaneshin/goenv)
    - You can install goenv by [anyenv](https://github.com/riywo/anyenv)
1. [Install the App Engine SDK for Go](https://cloud.google.com/appengine/docs/go/download?hl=ja)
1. `git clone git@github.com:groovenauts/blocks-concurrent-batch-agent.git`
1. [Install glide](https://github.com/Masterminds/glide#install)
1. `glide install`

## Run test

```
goapp test
```

### With coverage

```
goapp test -coverprofile coverage.out
go tool cover -html=coverage.out
```

## Run server locally

```
$ dev_appserver.py \
  ./app.yaml
```

### Client example

```
$ curl -c my.cookie 'http://localhost:8080/_ah/login?email=test%40example.com&admin=True&action=Login&continue='
$ curl -b my.cookie -X POST http://localhost:8080/pipelines.json --data '{"project_id":"FOO"}' -H 'Content-Type: application/json'
$ curl -b my.cookie http://localhost:8080/pipelines.json
```

## Deploy

```
$ appcfg.py \
  -A <YOUR_GCP_PROJECT> \
  -V $(cat VERSION) \
  update .
```

If you want to set it active, run the following command

```
$ gcloud app services set-traffic concurrent-batch-agent --splits=$(cat VERSION)=1
```

### New Pipeline data

1. Open the https://<hostname>/pipelines.html
2. Click [New Pipeline]
3. Fill in the fields
4. Click [submit]
