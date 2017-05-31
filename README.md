# blocks-concurrent-batch-agent

[![Build Status](https://secure.travis-ci.org/groovenauts/blocks-concurrent-batch-agent.png)](https://travis-ci.org/groovenauts/blocks-concurrent-batch-agent/)

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
$ make run
```

### Get Token on browser

1. Open http://localhost:8080/_ah/login and login
2. Open http://localhost:8080/admin/auths
3. Click [Create new token]
4. Copy the token shown

### Call API with curl

Make `pipeline.json` like this:

```json
{
  "name":"pipeline01",
  "project_id":"proj-123",
  "zone":"us-central1-f",
  "boot_disk": {
    "source_image":"https://www.googleapis.com/compute/v1/projects/google-containers/global/images/gci-stable-55-8872-76-0",
    "disk_type": "pd-ssd",
    "disk_size_gb": 30
  },
  "machine_type":"f1-micro",
  "preemptible": true,
  "target_size":1,
  "container_size":1,
  "container_name":"groovenauts/concurrent_batch_basic_example:0.4.0-alpha2",
  "command":""
}
```

```
$ TOKEN="[the token you got before]"
$ curl -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -X POST http://localhost:8080/pipelines --data @pipeline.json
$ curl -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' http://localhost:8080/pipelines
```

```
$ curl -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -X DELETE http://localhost:8080/pipelines/$ID
```

## Deploy to appengine

```
$ PROJECT=<YOUR_GCP_PROJECT> make deploy
```

If you want to set it active, run the following command

```
$ gcloud --project ${PROJECT} app services set-traffic concurrent-batch-agent --splits=$(cat VERSION)=1
```

### Get Token on browser

2. Open http://<hostname>/admin/auths
3. Click [Create new token]
4. Copy the token shown

### New Pipeline data


```
$ export TOKEN="[the token you got before]"
$ export AEHOST="[the host name you deployed]"
$ curl -v -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -X POST http://$AEHOST/pipelines --data @pipeline.json
```

#### Temporary work around

Now you have to call the following command to refresh status

```
$ curl -v -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' http://$AEHOST/pipelines/refresh
```

### Show all Pipeline data

$ curl -v -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' http://$AEHOST/pipelines

### Close and Delete data

```
$ export ID="[id of the result]"
$ curl -v -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -X PUT http://$AEHOST/pipelines/$ID/close --data ""
$ curl -v -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -X DELETE http://$AEHOST/pipelines/$ID
```
