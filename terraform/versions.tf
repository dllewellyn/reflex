terraform {
  required_version = ">= 1.0"
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = ">= 4.0"
    }
    random = {
      source  = "hashicorp/random"
      version = ">= 3.0"
    }
    pinecone = {
      source  = "pinecone-io/pinecone"
      version = ">=2.0.0"
    }
  }
}
