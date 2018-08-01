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

## Run test

```
make test
```

### With coverage

```
make test-coverage
open test/coverage.yyyy-mm-ddThh:mm:ssZ/index.html
```

## Run server locally

```
$ make run
```

### Get Token on browser

1. Open http://localhost:8080/_ah/login and login
2. Open http://localhost:8080/admin/orgs
3. Click [New Organization]
4. Enter your organization Name and click [Create]
5. Reload the page unless your organization appears
6. Click [Show] of your organization
7. Click [Auth List]
8. Click [Create new token]
9. Copy the organization ID and the token

### Call API with curl

Make `pipeline.json` like this:

```json
{
  "name":"pipeline01",
  "project_id":"proj-123",
  "zone":"us-central1-f",
  "boot_disk": {
    "source_image":"https://www.googleapis.com/compute/v1/projects/cos-cloud/global/images/family/cos-stable",
    "disk_type": "pd-ssd",
    "disk_size_gb": 30
  },
  "machine_type":"f1-micro",
  "preemptible": true,
  "stackdriver_agent": true,
  "target_size":1,
  "container_size":1,
  "container_name":"groovenauts/concurrent_batch_basic_example:0.4.0-alpha2",
  "command":"",
  "token_consumption": 1
}
```

```
$ export AEHOST=localhost:8080
$ export ORG_ID="[the organization ID you got before]"
$ export TOKEN="[the token you got before]"
$ curl -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -X POST http://$AEHOST/orgs/$ORG_ID/pipelines --data @pipeline.json
$ curl -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' http://$AEHOST/orgs/$ORG_ID/pipelines
```

```
$ curl -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -X DELETE http://$AEHOST/orgs/$ORG_ID/pipelines/$ID
```

## Deploy to appengine

```
$ export PROJECT=<YOUR_GCP_PROJECT>
$ make deploy
```

If you want to set it active, run the following command

```
$ make update-traffic
```

### Get Token on browser

`$AEHOST` means `the host name you deployed`

1. Open http://$AEHOST/admin/orgs
2. Click [New Organization]
3. Enter your organization Name and click [Create]
4. Reload the page unless your organization appears
5. Click [Show] of your organization
6. Click [Auth List]
7. Click [Create new token]
8. Copy the token shown

### New Pipeline data

```
$ export AEHOST="[the host name you deployed]"
$ export ORG_ID="[the organization ID you got before]"
$ export TOKEN="[the token you got before]"
$ curl -v -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -X POST http://$AEHOST/orgs/$ORG_ID/pipelines --data @pipeline.json
```


### Show all Pipeline data

$ curl -v -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' http://$AEHOST/orgs/$ORG_ID/pipelines


### Send a job message

Create job.json like this:

```json
{
  "id_by_client": "ID on your app",
  "message":{
    "attributes":{
      "download_files":"[\"gs://bucket1/path/to/file1\"]"
    }
  }
}
```

And publish (or reserve publishing) the job to the pipeline.

```
$ export ID="[id of the pipeline]"
$ curl -v -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -X POST "http://$AEHOST/pipelines/$ID/jobs?ready=true" --data @job.json
```




### Close and Delete data

```
$ export ID="[id of the result]"
$ curl -v -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -X PUT http://$AEHOST/pipelines/$ID/close --data ""
$ curl -v -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -X DELETE http://$AEHOST/pipelines/$ID
```

## Message

### pipeline.json

| Name                    | Type     | Required | Description   |
|-------------------------|----------|:--------:|---------------|
| boot_disk               | object   | true     | Boot disk for VM  |
| boot_disk.disk_size_gb  | int      | false    | Boot disk size in GB|
| boot_disk.disk_type     | string   | false    | Boot disk type: "pd-standard", "pd-ssd" |
| boot_disk.source_image  | string   | true     | Boot disk source image URL |
| close_policy            | int      | false    | Close policy at the end of jobs: 0: CloseAnyway, 1: CloseOnAllSuccess, 2: CloseNever |
| command                 | string   | false    | Command given to container |
| container_size          | int      | true     | The number of containers on each VM |
| container_name          | string   | true     | Container name to pull |
| dependency              | object   | false    | Dependency to jobs |
| dependency.condition    | int      | false    | Job's Condition to start the pipeline: 0:OnSuccess, 1:OnFailure, 2: OnFinish |
| dependency.job_ids      | []string | true     | Job IDs which the pipeline waits to finish |
| docker_run_options      | string   | false    | Options for `docker run` in startup script |
| gpu_accelerators        | object   | false    | GPU accelerator settings |
| gpu_accelerators.Count  | int      | true     | The number of GPU accelerators to use |
| gpu_accelerators.Type   | string   | true     | GPU accelerator type name (not URL). Run `gcloud compute accelerator-types list` |
| hibernation_delay       | int      | false    | The number of second to start hibernation after all of the jobs finished |
| job_scaler              | object   | false    | Setting to scale out  |
| job_scaler.enabled      | bool     | true     | If true, scaling out is enabled |
| job_scaler.max_instance_size | int | true     | Max instance size to increase by job_scaler |
| machine_type            | string   | true     | VM Machine type: Run `gcloud compute machine-types list` |
| name                    | string   | true     | Name of the pipeline |
| pulling                 | object   | false    | Pulling settings |
| pulling.message_per_pull | int     | false    | The number of messages to pull once. Default is 100. |
| pulling.interval_seconds | int     | false    | The number of second of interval to pull. Default is 30. |
| pulling.jobs_per_task    | int     | false    | The number of jobs to pull in a task. Default is 50. |
| preemptible             | bool     | false    | If true, use preemptible VMs |
| project_id              | string   | true     | GCP Project ID to run |
| stackdriver_agent       | bool     | false    | If true, use stackdriver agent |
| target_size             | int      | true     | The number of VMs |
| token_consumption       | int      | false    | The number of Organization tokens to consume |
| zone                    | string   | true     | GCP zone to run |

### job.json

| Name                    | Type              | Required | Description   |
|-------------------------|-------------------|:--------:|---------------|
| id_by_client            | string            | true     | The ID which client app generate for the job |
| message                 | map               | false    | The message to publish to pipeline-job-topic |
| message.attributes      | map[string]string | false    | The attributes of the message |
| message.data            | string            | false    | The data of the message |

#### Max size

The max size of key of `message.attributes` is 256 bytes.
The max size of value of `message.attributes` is 1,024 bytes.

#### use-data-as-attributes

If the data for `message.attributes` has the key or value which is more than each max size,
You can pass the data to `message.data` in JSON format by setting an attribute `use-data-as-attributes` `"true"`.
