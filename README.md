# clambucket

clambucket is a simple serverless virus scanning service for AWS S3. It uses the ClamAV virus scanner to scan files uploaded to an S3 bucket. It is designed to be deployed as a Kubernetes container, using SQS to trigger a scan when a new file is uploaded to the S3 bucket.

Files uploaded to an untrusted bucket will be scanned and if a virus is found, the file will be moved to a specified quarantine bucket. If no virus is found, the file will be moved to a specified clean bucket.

## Concepts

- Untrusted bucket: The S3 bucket where files are uploaded to be scanned. This is the bucket that will trigger the scan.
- Clean bucket: The S3 bucket where files are moved to if no virus is found.
- Quarantine bucket: The S3 bucket where files are moved to if a virus is found.
- Event handler: The trigger which will notify clambucket that a new file has been uploaded to the untrusted bucket. When launched, clambucket will connect to the event handler to retrieve the file to scan.

## Deployment

clambucket can be deployed in a couple different ways. In both models, clambucket relies on either the environment or cloud provider to inject the appropriate credentials to access the S3 bucket and SQS queue.

### Configuration

First, you need to create a secret/configmap to contain the configuration values for clambucket. You can reference the provided `.env-sample` file.

```bash
ASSUME_ROLE_ARN= # optional. role to assume before calling AWS APIs
CLEAN_BUCKET= # required. the S3 bucket where clean files will be moved to.
CLEAN_PREFIX= # optional. the prefix to move clean files to within the clean bucket.
EVENT_CONFIG= # optional. the configuration for the event handler. varies by event handler.
EVENT_HANDLER= # required. the event handler to use. currently only supports "sqs" and "cli".
LOG_LEVEL= # optional. the log level to use. defaults to "info".
QUARANTINE_BUCKET= # required. the S3 bucket where quarantined files will be moved to.
QUARANTINE_PREFIX= # optional. the prefix to move quarantined files to within the quarantine bucket.
WATCH= # optional. run clambucket in watch mode. defaults to "false".
```

Note that using the prefixes, you can actually move files to a different directory within the same bucket. However if doing this, take caution to ensure your S3 event trigger is configured to NOT trigger on the files being moved to the clean or quarantine prefixes, or you will end up in an infinite loop.

```bash
kubectl create secret generic clambucket --from-env-file=.env
```

### Identity and Access

clambucket uses a Kubernetes Service Account which can be bound to an IAM Role. Follow [this guide](https://docs.aws.amazon.com/AmazonS3/latest/userguide/ways-to-add-notification-config-to-bucket.html#step1-create-sqs-queue-for-notification) to configure your bucket and queue appropriately.

```bash
kubectl apply -f manifests/serviceaccount.yaml
```

### ClamAV Database

clambucket requires the ClamAV database to be available. clambucket uses a `RWX` PVC to persist the database between pod restarts, and a `CronJob` to update the database daily.

```bash
kubectl apply -f manifests/pvc.yaml
kubectl apply -f manifests/database-updater.yaml
```

### Keda ScaledJob

clambucket can be deployed as a Keda ScaledJob. This is the recommended deployment model.

You can create a SQS queue and set up S3 event notifcations on `ObjectCreated:*` events to send a message to the SQS queue. Then, you can deploy clambucket as a Keda ScaledJob monitoring this queue. When a new message is sent to the queue, Keda will scale up the clambucket job to process the message, and then scale it back down when the job is complete. In this model, one clambucket pod will process one file at a time, but scale out horizontally to handle multiple files concurrently.

Ensure you've configured Keda appropriately to access the SQS queue, then deploy the scaled job.

```bash
kubectl apply -f manifests/scaled-job
```

### Kubernetes Deployment

clambucket can also be deployed as a Kubernetes Deployment. This is not recommended, as it will not scale out to handle multiple files concurrently, but rather process through files consecutively as they are uploaded. In this model, clambucket is deployed as a long-running process, and will process files as they are uploaded to the untrusted bucket.

```bash
kubectl apply -f manifests/deployment
```