resource "google_artifact_registry_repository" "backend" {
  location      = var.region
  repository_id = "kctfestnav"
  format        = "DOCKER"
  # 新しい 5 つを残し、1 週間より古いものは消す
  cleanup_policies {
    id     = "keep-recent"
    action = "KEEP"
    most_recent_versions { keep_count = 5 }
  }
  cleanup_policies {
    id     = "delete-old"
    action = "DELETE"
    condition { older_than = "604800s" }
  }
  depends_on = [google_project_service.apis]
}
