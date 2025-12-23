# Ingestor Service (Cloud Run)
resource "google_cloud_run_v2_service" "ingestor" {
  provider = google
  name     = "ingestor"
  location = var.region
  project  = var.project_id
  ingress  = "INGRESS_TRAFFIC_ALL"

  template {
    timeout         = "300s"
    service_account = google_service_account.ingestor_sa.email

    containers {
      image = "${var.region}-docker.pkg.dev/${var.project_id}/${google_artifact_registry_repository.flash_cache_repo.repository_id}/ingestor:latest"


      env {
        name  = "KAFKA_TOPIC"
        value = var.kafka_topic
      }
      env {
        name  = "KAFKA_BOOTSTRAP_SERVERS"
        value = var.kafka_bootstrap_servers
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
        name  = "GOOGLE_CLOUD_PROJECT"
        value = var.project_id
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

  traffic {
    type    = "TRAFFIC_TARGET_ALLOCATION_TYPE_LATEST"
    percent = 100
  }
}

# Service Account for Ingestor
resource "google_service_account" "ingestor_sa" {
  provider     = google
  account_id   = "flash-cache-ingestor-sa"
  display_name = "Service Account for Reflex Ingestor"
}

# Grant Cloud Trace Agent role
resource "google_project_iam_member" "ingestor_trace_binding" {
  provider = google
  project  = var.project_id
  role     = "roles/cloudtrace.agent"
  member   = "serviceAccount:${google_service_account.ingestor_sa.email}"
}

# Grant Monitoring Metric Writer role
resource "google_project_iam_member" "ingestor_metric_binding" {
  provider = google
  project  = var.project_id
  role     = "roles/monitoring.metricWriter"
  member   = "serviceAccount:${google_service_account.ingestor_sa.email}"
}

# Allow unauthenticated invocations (Public API)
resource "google_cloud_run_service_iam_member" "ingestor_public_access" {
  location = google_cloud_run_v2_service.ingestor.location
  project  = google_cloud_run_v2_service.ingestor.project
  service  = google_cloud_run_v2_service.ingestor.name
  role     = "roles/run.invoker"
  member   = "allUsers"
}

output "ingestor_url" {
  value = google_cloud_run_v2_service.ingestor.uri
}
