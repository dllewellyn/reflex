output "raw_prompts_bucket" {
  value       = google_storage_bucket.raw_prompts.name
  description = "The name of the raw prompts bucket"
}

output "batch_staging_bucket" {
  value       = google_storage_bucket.batch_staging.name
  description = "The name of the batch staging bucket"
}

output "processed_prompts_bucket" {
  value       = google_storage_bucket.processed_prompts.name
  description = "The name of the processed prompts bucket"
}

output "project_id" {
  value       = var.project_id
  description = "The project ID"
}

output "region" {
  value       = var.region
  description = "The region"
}
