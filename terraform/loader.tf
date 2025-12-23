# Service Account for the hourly loader job (Kafka to GCS)
resource "google_service_account" "loader_sa" {
  provider     = google
  account_id   = "flash-cache-loader-sa"
  display_name = "Service Account for Reflex Loader (Kafka to GCS)"
}

# Grant GCS Object Admin role to loader service account
resource "google_project_iam_member" "loader_gcs_admin_binding" {
  provider = google
  project  = var.project_id
  role     = "roles/storage.objectAdmin"
  member   = "serviceAccount:${google_service_account.loader_sa.email}"
}

# Grant Cloud Trace Agent role to loader
resource "google_project_iam_member" "loader_trace_binding" {
  provider = google
  project  = var.project_id
  role     = "roles/cloudtrace.agent"
  member   = "serviceAccount:${google_service_account.loader_sa.email}"
}

# Grant Monitoring Metric Writer role to loader
resource "google_project_iam_member" "loader_metric_binding" {
  provider = google
  project  = var.project_id
  role     = "roles/monitoring.metricWriter"
  member   = "serviceAccount:${google_service_account.loader_sa.email}"
}

# Grant Cloud Run Invoker role to loader (Required for Cloud Scheduler)
resource "google_project_iam_member" "loader_run_invoker_binding" {
  provider = google
  project  = var.project_id
  role     = "roles/run.invoker"
  member   = "serviceAccount:${google_service_account.loader_sa.email}"
}

# Cloud Run Job: interaction-loader (Hourly Kafka to GCS)
resource "google_cloud_run_v2_job" "interaction_loader" {
  provider            = google
  name                = "interaction-loader"
  location            = var.region
  project             = var.project_id
  deletion_protection = false

  template {
    template {
      service_account = google_service_account.loader_sa.email
      containers {
        image = "${var.region}-docker.pkg.dev/${var.project_id}/${google_artifact_registry_repository.flash_cache_repo.repository_id}/loader:latest"
        env {
          name  = "GCP_PROJECT_ID"
          value = var.project_id
        }
        env {
          name  = "GOOGLE_CLOUD_PROJECT"
          value = var.project_id
        }
        env {
          name  = "GCS_RAW_PROMPT_BUCKET"
          value = google_storage_bucket.raw_prompts.name
        }
        # KAFKA_BROKERS and KAFKA_TOPIC need to be passed here.
        env {
          name  = "KAFKA_BOOTSTRAP_SERVERS"
          value = var.kafka_bootstrap_servers
        }
        env {
          name  = "KAFKA_TOPIC"
          value = var.kafka_topic
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
          value = "gcs-loader-v1"
        }
      }
    }
  }
}

# Cloud Scheduler Job: trigger-loader-hourly
resource "google_cloud_scheduler_job" "trigger_loader_hourly" {
  provider = google
  name     = "trigger-loader-hourly"
  region   = var.region
  project  = var.project_id
  schedule = "0 * * * *" # Every hour

  http_target {
    uri         = "https://run.googleapis.com/v2/projects/${var.project_id}/locations/${var.region}/jobs/${google_cloud_run_v2_job.interaction_loader.name}:run"
    http_method = "POST"
    oauth_token {
      service_account_email = google_service_account.loader_sa.email
      scope                 = "https://www.googleapis.com/auth/cloud-platform"
    }
  }
}
