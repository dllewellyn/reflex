resource "pinecone_index" "flash_cache" {
  name      = "flash-cache"
  dimension = 1024
  metric    = "cosine"

  tags = {
    environment = "dev"
  }

  embed = {
      model = "multilingual-e5-large"
      field_map = {
          text = "chunk_text"
      }
  }

  spec = {
    serverless = {
      cloud  = "aws"
      region = "us-east-1"
    }
  }
}

output "pinecone_index_host" {
  value = pinecone_index.flash_cache.host
}