variable "project_id" {
  description = "The ID of the Google Cloud project"
  type        = string
}

variable "region" {
  description = "The region to deploy resources in"
  type        = string
  default     = "us-central1"
}

variable "bucket_name" {
  description = "Name of the GCS bucket for raw conversation history"
  type        = string
  default     = "flash-cache-conversation-history-bronze"
}

variable "kafka_bootstrap_servers" {
  description = "Kafka Bootstrap Servers"
  type        = string
}

variable "kafka_topic" {
  description = "Kafka Topic to ingest to"
  type        = string
}

variable "kafka_api_key" {
  description = "Kafka API Key"
  type        = string
  sensitive   = true
}

variable "kafka_api_secret" {
  description = "Kafka API Secret"
  type        = string
  sensitive   = true
}

variable "kafka_batch_results_topic" {
  description = "Kafka Topic for batch results"
  type        = string
  default     = "batch-job-results"
}

variable "pinecone_api_key" {
  description = "Pinecone API Key"
  type        = string
  sensitive   = true
}


