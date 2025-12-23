# GCS Log Ingestion Setup Guide

This guide provides executable commands to set up test environments for both usage logs and audit logs.

> [!NOTE]
> Ensure you are authenticated with `gcloud auth login` and have a valid project selected.

## Common Setup
Set environment variables for uniqueness.
```bash
export PROJECT_ID=$(gcloud config get-value project)
export RANDOM_SUFFIX=$(openssl rand -hex 4)
export SOURCE_BUCKET="test-source-${PROJECT_ID}-${RANDOM_SUFFIX}"
export LOGS_BUCKET="test-logs-${PROJECT_ID}-${RANDOM_SUFFIX}"
export AUDIT_BUCKET="test-audit-${PROJECT_ID}-${RANDOM_SUFFIX}"

echo "Source Bucket: $SOURCE_BUCKET"
echo "Logs Bucket:   $LOGS_BUCKET"
echo "Audit Bucket:  $AUDIT_BUCKET"
```

---

## Scenario A: Usage Logs (Classic CSV)
Classic hourly usage logs.

### 1. Create Buckets
```bash
gcloud storage buckets create gs://$SOURCE_BUCKET --location=US
gcloud storage buckets create gs://$LOGS_BUCKET --location=US
```

### 2. Enable Logging
```bash
gsutil logging set on -b gs://$LOGS_BUCKET gs://$SOURCE_BUCKET
```

### 3. Generate Traffic
```bash
echo "Hello GCS" > hello.txt
gcloud storage cp hello.txt gs://$SOURCE_BUCKET/hello.txt
gcloud storage cat gs://$SOURCE_BUCKET/hello.txt
```

### 4. Wait & Run (Requires ~1hr wait)
Usage logs take about an hour to appear. Once they show up in `gs://$LOGS_BUCKET`, run:
```bash
# Find the latest log file
LOG_FILE=$(gcloud storage ls "gs://$LOGS_BUCKET/${SOURCE_BUCKET}_usage*" | head -n 1)

# Run Ingestor
go run main.go --source "$LOG_FILE" --format usage
```

---

## Scenario B: Cloud Audit Logs (Modern JSON)
Near real-time JSON logs via Cloud Logging Sink.

### 1. Create Audit Log Bucket
```bash
gcloud storage buckets create gs://$AUDIT_BUCKET --location=US
```

### 2. Create Log Sink
Creates a sink to export Object Data Access logs to the bucket.
```bash
# Grant bucket write permission to the logging service identity will be handled automatically by gcloud in most cases, 
# but if not, notice the service account in the output and grant it roles/storage.objectCreator.

gcloud logging sinks create "gcs-audit-sink-${RANDOM_SUFFIX}" \
    storage.googleapis.com/$AUDIT_BUCKET \
    --log-filter='resource.type="gcs_bucket" AND protoPayload.methodName:"storage.objects."' \
    --description="Export GCS data access logs to bucket"
```

### 3. Generate Traffic
Using the same source bucket from Scenario A (or create a new one).
```bash
# Ensure Access Logs are enabled for the bucket if they aren't by default (Data Access logs)
# (In generic setup, Data Access logs might be disabled by default. We enable them here.)
# Note: This enables it for the whole project or bucket. Let's do bucket level if possible or rely on project defaults.
# simpler to just assume they are on or enable them project wide:
# gcloud projects get-iam-policy $PROJECT_ID ... (complex)
# For PoC, let's assume default audit config or that user has permissions.
```

### 4. Wait & Run (Seconds to Minutes)
Audit logs appear much faster.
```bash
# Wait a minute, then find the latest log file
# Logs are partitioned: gs://BUCKET/cloudaudit.googleapis.com/data_access/YYYY/MM/DD/...
AUDIT_LOG_FILE=$(gcloud storage ls -r "gs://$AUDIT_BUCKET/**.json" | tail -n 1)

# Run Ingestor
go run main.go --source "$AUDIT_LOG_FILE" --format audit
```

---

## Cleanup
Remove the test resources.
```bash
# Delete Buckets
gcloud storage rm -r gs://$SOURCE_BUCKET
gcloud storage rm -r gs://$LOGS_BUCKET
gcloud storage rm -r gs://$AUDIT_BUCKET

# Delete Sink
gcloud logging sinks delete "gcs-audit-sink-${RANDOM_SUFFIX}"
```
