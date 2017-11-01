# Capped pipelines test

## Setup

Open [http://localhost:8080/admin/orgs](http://localhost:8080/admin/orgs).
Set `3` to the `TokenAmount` of the target organization.

```bash
$ cd test/pipeline_cap
$ export AEHOST="[the host name you deployed]"
$ export ORG_ID="[the organization ID you got before]"
$ export TOKEN="[the token you got before]"
```

```bash
$ export PIPELINE1_ID=$(curl -v -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -X POST http://$AEHOST/orgs/$ORG_ID/pipelines --data @pipeline1.json | jq -r ".id")
$ curl -v -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -X POST http://$AEHOST/pipelines/$PIPELINE1_ID/jobs --data @job.json
$ curl -v -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -X POST http://$AEHOST/pipelines/$PIPELINE1_ID/jobs --data @job.json
```

```bash
$ export PIPELINE2_ID=$(curl -v -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -X POST http://$AEHOST/orgs/$ORG_ID/pipelines --data @pipeline2.json | jq -r ".id")
$ curl -v -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -X POST http://$AEHOST/pipelines/$PIPELINE2_ID/jobs --data @job.json
$ curl -v -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -X POST http://$AEHOST/pipelines/$PIPELINE2_ID/jobs --data @job.json
$ curl -v -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -X POST http://$AEHOST/pipelines/$PIPELINE2_ID/jobs --data @job.json
```

```bash
$ export PIPELINE3_ID=$(curl -v -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -X POST http://$AEHOST/orgs/$ORG_ID/pipelines --data @pipeline3.json | jq -r ".id")
$ curl -v -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -X POST http://$AEHOST/pipelines/$PIPELINE3_ID/jobs --data @job.json
$ curl -v -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -X POST http://$AEHOST/pipelines/$PIPELINE3_ID/jobs --data @job.json
$ curl -v -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -X POST http://$AEHOST/pipelines/$PIPELINE3_ID/jobs --data @job.json
```

```bash
$ curl -v -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' http://$AEHOST/pipelines/refresh
```

Then `pipeline1` starts working but `pipeline2` and `pipeline3` don't start working.

```bash
$ pubsub-devsub --project dummy-project-999 --subscription akm-pipeline01-progress-subscription
```

After `pipeline1` jobs finish, close and delete `pipeline1`.

```bash
$ curl -v -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -X PUT http://$AEHOST/pipelines/$PIPELINE1_ID/close --data ""
```

Repeat the following until the pipeline's status will be 9.

```bash
$ curl -v -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' http://$AEHOST/pipelines/refresh
$ curl -v -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' http://$AEHOST/orgs/$ORG_ID/pipelines | jq .
```

And check if pipeline2 is starting and `pipeline3` isn't starting.

```bash
$ pubsub-devsub --project dummy-project-999 --subscription akm-pipeline02-progress-subscription
```


And delete pipeline1.

```bash
$ curl -v -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -X DELETE http://$AEHOST/pipelines/$PIPELINE1_ID
```


After `pipeline2` jobs finish, close and delete `pipeline2`.

```bash
$ curl -v -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -X PUT http://$AEHOST/pipelines/$PIPELINE2_ID/close --data ""
```

Repeat the following until the pipeline's status will be 9.

```bash
$ curl -v -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' http://$AEHOST/pipelines/refresh
$ curl -v -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' http://$AEHOST/orgs/$ORG_ID/pipelines | jq .
```


And check if pipeline3 is starting.

```bash
$ pubsub-devsub --project dummy-project-999 --subscription akm-pipeline03-progress-subscription
```

And delete pipeline2.

```bash
$ curl -v -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -X DELETE http://$AEHOST/pipelines/$PIPELINE2_ID
```


After `pipeline3` jobs finish, close and delete `pipeline3`.

```bash
$ curl -v -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -X PUT http://$AEHOST/pipelines/$PIPELINE3_ID/close --data ""
```

Repeat the following until the pipeline's status will be 9.

```bash
$ curl -v -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' http://$AEHOST/pipelines/refresh
$ curl -v -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' http://$AEHOST/orgs/$ORG_ID/pipelines | jq .
```

And delete pipeline3.

```bash
$ curl -v -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -X DELETE http://$AEHOST/pipelines/$PIPELINE3_ID
```
