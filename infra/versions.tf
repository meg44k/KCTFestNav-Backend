terraform {
  required_version = ">= 1.9"
  required_providers {
    google     = { source = "hashicorp/google", version = "~> 8.0" }
    cloudflare = { source = "cloudflare/cloudflare", version = "~> 5.0" }
    random     = { source = "hashicorp/random", version = "~> 3.6" }
  }
  # バケット名は terraform init -backend-config="bucket=..." で渡す
  backend "gcs" {
    prefix = "kctfestnav"
  }
}
