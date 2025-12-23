#!/bin/bash
set -e

# Load .env variables
if [ -f .env ]; then
  export $(grep -v '^#' .env | xargs)
else
  echo ".env file not found!"
  exit 1
fi

if [ -z "$GOOGLE_CLOUD_PROJECT" ]; then
  echo "GOOGLE_CLOUD_PROJECT is not set in .env"
  exit 1
fi

BUCKET_NAME="${GOOGLE_CLOUD_PROJECT}-tf-state"
REGION="us-central1" # Default or read from env if available

echo "Setting up infrastructure for project: $GOOGLE_CLOUD_PROJECT"
echo "State bucket name: $BUCKET_NAME"

# Check if bucket exists
if ! gcloud storage buckets describe "gs://${BUCKET_NAME}" &>/dev/null; then
  echo "Creating state bucket..."
  gcloud storage buckets create "gs://${BUCKET_NAME}" --project="$GOOGLE_CLOUD_PROJECT" --location="$REGION" --uniform-bucket-level-access
else
  echo "State bucket already exists."
fi

# Ensure versioning is enabled (idempotent)
gcloud storage buckets update "gs://${BUCKET_NAME}" --versioning

# Generate backend.tf
echo "Generating terraform/backend.tf..."
cat > terraform/backend.tf <<EOF
terraform {
  backend "gcs" {
    bucket = "${BUCKET_NAME}"
    prefix = "hub"
  }
}
EOF

echo "Setup complete. You can now run 'make init' and 'make deploy-hub'."
