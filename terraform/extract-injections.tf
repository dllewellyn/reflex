# Service Account for Extract Injections Job
resource "google_service_account" "extract_injections_sa" {
  provider     = google
  account_id   = "flash-cache-extract-sa"
  display_name = "Service Account for Reflex Extract Injections"
}

# Grant Cloud Trace Agent role
resource "google_project_iam_member" "extract_trace_binding" {
  provider = google
  project  = var.project_id
  role     = "roles/cloudtrace.agent"
  member   = "serviceAccount:${google_service_account.extract_injections_sa.email}"
}

# Grant Monitoring Metric Writer role
resource "google_project_iam_member" "extract_metric_binding" {
  provider = google
  project  = var.project_id
  role     = "roles/monitoring.metricWriter"
  member   = "serviceAccount:${google_service_account.extract_injections_sa.email}"
}

# Grant Vertex AI User role to extract injections service account
resource "google_project_iam_member" "extract_vertex_user_binding" {
  provider = google
  project  = var.project_id
  role     = "roles/aiplatform.user"
  member   = "serviceAccount:${google_service_account.extract_injections_sa.email}"
}

# Grant Cloud Run Invoker role to extract-injections (Required for Cloud Scheduler)
resource "google_project_iam_member" "extract_run_invoker_binding" {
  provider = google
  project  = var.project_id
  role     = "roles/run.invoker"
  member   = "serviceAccount:${google_service_account.extract_injections_sa.email}"
}

# Cloud Run Job: extract-injections
resource "google_cloud_run_v2_job" "extract_injections" {
  provider            = google
  name                = "extract-injections"
  location            = var.region
  project             = var.project_id
  deletion_protection = false


  template {
    template {
      timeout = "3600s"

      service_account = google_service_account.extract_injections_sa.email
      containers {
        image = "${var.region}-docker.pkg.dev/${var.project_id}/${google_artifact_registry_repository.flash_cache_repo.repository_id}/extract-injections:latest"

        env {
          name  = "GOOGLE_CLOUD_PROJECT"
          value = var.project_id
        }
        env {
          name  = "GCP_PROJECT_ID"
          value = var.project_id
        }
        env {
          name  = "KAFKA_BOOTSTRAP_SERVERS"
          value = var.kafka_bootstrap_servers
        }
        env {
          name  = "KAFKA_TOPIC_BATCH_RESULTS"
          value = var.kafka_batch_results_topic
        }
        env {
          name  = "KAFKA_API_KEY"
          value = var.kafka_api_key
        }
        env {
          name  = "KAFKA_API_SECRET"
          value = var.kafka_api_secret
        }
        env {
          name  = "KAFKA_CONSUMER_GROUP_ID"
          value = "extract-injections-v1"
        }
        env {
          name  = "PINECONE_API_KEY"
          value = var.pinecone_api_key
        }
        env {
          name  = "PINECONE_INDEX_HOST"
          value = pinecone_index.flash_cache.host
        }
      }
    }
  }
}

# Cloud Scheduler Job: trigger-extract-injections
resource "google_cloud_scheduler_job" "trigger_extract_injections" {
  provider = google
  name     = "trigger-extract-injections"
  region   = var.region
  project  = var.project_id
  schedule = "0 */6 * * *" # Every 6 hours

  http_target {
    uri         = "https://run.googleapis.com/v2/projects/${var.project_id}/locations/${var.region}/jobs/${google_cloud_run_v2_job.extract_injections.name}:run"
    http_method = "POST"
    oauth_token {
      service_account_email = google_service_account.extract_injections_sa.email
      scope                 = "https://www.googleapis.com/auth/cloud-platform"
    }
  }
}
