terraform {
  backend "gcs" {
    bucket = "flash-cache-poc-tf-state"
    prefix = "hub"
  }
}
