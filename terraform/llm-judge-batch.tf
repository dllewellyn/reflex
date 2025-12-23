resource "google_artifact_registry_repository" "flash_cache_repo" {
  provider      = google
  repository_id = "flash-cache-repo"
  description   = "Docker repository for flash-cache application images"
  format        = "DOCKER"
  location      = var.region
}

# Service Account for the hourly loader job (Kafka to GCS)


# Service Account for the daily LLM Judge batch job (GCS to Vertex AI)
resource "google_service_account" "batch_judge_sa" {
  provider     = google
  account_id   = "flash-cache-batch-judge-sa"
  display_name = "Service Account for Reflex LLM Judge Batch"
}

# Grant GCS Object Admin role to batch judge service account
resource "google_project_iam_member" "batch_judge_gcs_admin_binding" {
  provider = google
  project  = var.project_id
  role     = "roles/storage.objectAdmin"
  member   = "serviceAccount:${google_service_account.batch_judge_sa.email}"
}

# Grant Vertex AI User role to batch judge service account
resource "google_project_iam_member" "batch_judge_vertex_user_binding" {
  provider = google
  project  = var.project_id
  role     = "roles/aiplatform.user"
  member   = "serviceAccount:${google_service_account.batch_judge_sa.email}"
}

# Grant Cloud Trace Agent role to loader


# Grant Cloud Trace Agent role to batch judge
resource "google_project_iam_member" "batch_judge_trace_binding" {
  provider = google
  project  = var.project_id
  role     = "roles/cloudtrace.agent"
  member   = "serviceAccount:${google_service_account.batch_judge_sa.email}"
}

# Grant Monitoring Metric Writer role to batch judge
resource "google_project_iam_member" "batch_judge_metric_binding" {
  provider = google
  project  = var.project_id
  role     = "roles/monitoring.metricWriter"
  member   = "serviceAccount:${google_service_account.batch_judge_sa.email}"
}

# Grant Cloud Run Invoker role to batch judge (Required for Cloud Scheduler)
resource "google_project_iam_member" "batch_judge_run_invoker_binding" {
  provider = google
  project  = var.project_id
  role     = "roles/run.invoker"
  member   = "serviceAccount:${google_service_account.batch_judge_sa.email}"
}

# Cloud Run Job: interaction-loader (Hourly Kafka to GCS)


# Cloud Run Job: llm-judge-trigger (Daily GCS to Vertex AI)
resource "google_cloud_run_v2_job" "llm_judge_trigger" {
  provider            = google
  name                = "llm-judge-trigger"
  location            = var.region
  project             = var.project_id
  deletion_protection = false

  template {
    template {
      service_account = google_service_account.batch_judge_sa.email
      containers {
        image = "${var.region}-docker.pkg.dev/${var.project_id}/${google_artifact_registry_repository.flash_cache_repo.repository_id}/batch-job:latest"
        env {
          name  = "GCP_PROJECT"
          value = var.project_id
        }
        env {
          name  = "GOOGLE_CLOUD_PROJECT"
          value = var.project_id
        }
        env {
          name  = "GCP_LOCATION"
          value = var.region
        }
        env {
          name  = "GCS_RAW_PROMPT_BUCKET"
          value = google_storage_bucket.raw_prompts.name
        }
        env {
          name  = "GCS_BATCH_STAGING_BUCKET"
          value = google_storage_bucket.batch_staging.name
        }
        env {
          name  = "GCS_PROCESSED_PROMPT_BUCKET"
          value = google_storage_bucket.processed_prompts.name
        }
        env {
          name  = "MODEL_ID"
          value = "publishers/google/models/gemini-2.5-flash"
        }
        env {
          name  = "PROMPT_PATH"
          value = "/prompts/security-judge.prompt.yml"
        }
      }
    }
  }
}

# Cloud Scheduler Job: trigger-judge-daily
resource "google_cloud_scheduler_job" "trigger_judge_daily" {
  provider = google
  name     = "trigger-judge-daily"
  region   = var.region
  project  = var.project_id
  schedule = "0 2 * * *" # 2:00 AM Daily

  http_target {
    uri         = "https://run.googleapis.com/v2/projects/${var.project_id}/locations/${var.region}/jobs/${google_cloud_run_v2_job.llm_judge_trigger.name}:run"
    http_method = "POST"
    oauth_token {
      service_account_email = google_service_account.batch_judge_sa.email
      scope                 = "https://www.googleapis.com/auth/cloud-platform"
    }
  }
}

# --------------------------------------------------------------------------------------------------
# Batch Result Processing Automation (Cloud Function)
# --------------------------------------------------------------------------------------------------

# Service Account for the function
resource "google_service_account" "batch_result_processor_sa" {
  provider     = google
  account_id   = "batch-result-processor-sa"
  display_name = "Service Account for Batch Result Processor Function"
}

# Grant Storage Object Viewer on the processed prompts bucket (Trigger Source)
resource "google_storage_bucket_iam_member" "processor_gcs_viewer" {
  provider = google
  bucket   = google_storage_bucket.processed_prompts.name
  role     = "roles/storage.objectViewer"
  member   = "serviceAccount:${google_service_account.batch_result_processor_sa.email}"
}

# Grant Eventarc Event Receiver role
resource "google_project_iam_member" "processor_event_receiver" {
  provider = google
  project  = var.project_id
  role     = "roles/eventarc.eventReceiver"
  member   = "serviceAccount:${google_service_account.batch_result_processor_sa.email}"
}

# Grant Run Invoker role (Function Gen2 runs on Cloud Run)
resource "google_project_iam_member" "processor_run_invoker" {
  provider = google
  project  = var.project_id
  role     = "roles/run.invoker"
  member   = "serviceAccount:${google_service_account.batch_result_processor_sa.email}"
}

# Grant Cloud Trace Agent role
resource "google_project_iam_member" "processor_trace_binding" {
  provider = google
  project  = var.project_id
  role     = "roles/cloudtrace.agent"
  member   = "serviceAccount:${google_service_account.batch_result_processor_sa.email}"
}

# Grant Monitoring Metric Writer role
resource "google_project_iam_member" "processor_metric_binding" {
  provider = google
  project  = var.project_id
  role     = "roles/monitoring.metricWriter"
  member   = "serviceAccount:${google_service_account.batch_result_processor_sa.email}"
}

# Grant Pub/Sub Publisher to GCS Service Account (Required for Eventarc)
data "google_storage_project_service_account" "gcs_account" {
  provider = google
}

resource "google_project_iam_member" "gcs_pubsub_publisher" {
  provider = google
  project  = var.project_id
  role     = "roles/pubsub.publisher"
  member   = "serviceAccount:${data.google_storage_project_service_account.gcs_account.email_address}"
}

# Bucket for storing function source code
resource "google_storage_bucket" "function_source" {
  provider      = google
  name          = "${var.project_id}-function-source"
  location      = var.region
  force_destroy = true
}

# Zip the source code
data "archive_file" "function_source_zip" {
  type        = "zip"
  source_dir  = "../bin/function-source"
  output_path = "/tmp/batch-result-trigger.zip"
}

# Upload source to bucket
resource "google_storage_bucket_object" "function_source_object" {
  name   = "batch-result-trigger-${data.archive_file.function_source_zip.output_md5}.zip"
  bucket = google_storage_bucket.function_source.name
  source = data.archive_file.function_source_zip.output_path
}

# Cloud Function (Gen 2)
resource "google_cloudfunctions2_function" "batch_result_processor" {
  provider    = google
  name        = "batch-result-processor"
  location    = var.region
  description = "Processes new batch result files and streams to Kafka"

  build_config {
    runtime     = "go123"
    entry_point = "ProcessBatchResult"
    source {
      storage_source {
        bucket = google_storage_bucket.function_source.name
        object = google_storage_bucket_object.function_source_object.name
      }
    }
  }

  service_config {
    max_instance_count    = 10
    available_memory      = "512Mi"
    timeout_seconds       = 540
    service_account_email = google_service_account.batch_result_processor_sa.email

    environment_variables = {
      KAFKA_BOOTSTRAP_SERVERS = var.kafka_bootstrap_servers
      KAFKA_TOPIC             = var.kafka_batch_results_topic
      KAFKA_API_KEY           = var.kafka_api_key
      KAFKA_API_SECRET        = var.kafka_api_secret
    }
  }

  event_trigger {
    trigger_region = var.region
    event_type     = "google.cloud.storage.object.v1.finalized"
    event_filters {
      attribute = "bucket"
      value     = google_storage_bucket.processed_prompts.name
    }
    retry_policy = "RETRY_POLICY_DO_NOT_RETRY"
  }
}
