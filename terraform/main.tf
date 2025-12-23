provider "google" {
  project = var.project_id
  region  = var.region
}

provider "pinecone" {
  api_key = var.pinecone_api_key
}

resource "random_id" "bucket_suffix" {
  byte_length = 4
}

# Bucket 1: Raw Prompts (Ingestion Sink)
resource "google_storage_bucket" "raw_prompts" {
  name          = "${var.bucket_name}-raw-${random_id.bucket_suffix.hex}"
  location      = var.region
  force_destroy = false

  uniform_bucket_level_access = true

  versioning {
    enabled = true
  }
}

# Bucket 2: Batch Staging (Formatted for Vertex AI)
resource "google_storage_bucket" "batch_staging" {
  name          = "${var.bucket_name}-staging-${random_id.bucket_suffix.hex}"
  location      = var.region
  force_destroy = false

  uniform_bucket_level_access = true

  versioning {
    enabled = true
  }
}

# Bucket 3: Processed Prompts (Analysis Results)
resource "google_storage_bucket" "processed_prompts" {
  name          = "${var.bucket_name}-processed-${random_id.bucket_suffix.hex}"
  location      = var.region
  force_destroy = false

  uniform_bucket_level_access = true

  versioning {
    enabled = true
  }
}
