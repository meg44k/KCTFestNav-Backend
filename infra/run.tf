resource "google_service_account" "run" {
  account_id   = "kctfestnav-run"
  display_name = "KCTFestNav API (Cloud Run)"
}

resource "google_project_iam_member" "run_sql" {
  project = var.project_id
  role    = "roles/cloudsql.client"
  member  = "serviceAccount:${google_service_account.run.email}"
}

locals {
  plain_env = {
    DB_USER           = google_sql_user.app.name
    DB_NAME           = google_sql_database.app.name
    DB_SOCKET         = "/cloudsql/${google_sql_database_instance.main.connection_name}"
    DB_MAX_OPEN_CONNS = "5"
    REDIS_ADDR        = var.redis_addr
    REDIS_TLS         = "true"
    INIT_ADMIN_ID     = var.init_admin_id
  }
}

resource "google_cloud_run_v2_service" "api" {
  name                = "kctfestnav-api"
  location            = var.region
  ingress             = "INGRESS_TRAFFIC_ALL"
  deletion_protection = false

  template {
    service_account                  = google_service_account.run.email
    max_instance_request_concurrency = 80
    scaling {
      min_instance_count = var.min_instances
      max_instance_count = var.max_instances
    }
    volumes {
      name = "cloudsql"
      cloud_sql_instance {
        instances = [google_sql_database_instance.main.connection_name]
      }
    }
    containers {
      # 最初だけの仮のイメージ。以後は GitHub Actions が出す
      image = "us-docker.pkg.dev/cloudrun/container/hello"
      ports {
        container_port = 8080
      }
      resources {
        limits   = { cpu = "1", memory = "512Mi" }
        cpu_idle = true
      }
      volume_mounts {
        name       = "cloudsql"
        mount_path = "/cloudsql"
      }
      dynamic "env" {
        for_each = local.plain_env
        content {
          name  = env.key
          value = env.value
        }
      }
      dynamic "env" {
        for_each = toset(local.all_secrets)
        content {
          name = env.key
          value_source {
            secret_key_ref {
              secret  = google_secret_manager_secret.s[env.key].secret_id
              version = "latest"
            }
          }
        }
      }
    }
  }

  # Actions が出したイメージを terraform apply で hello に戻さない
  lifecycle {
    ignore_changes = [
      template[0].containers[0].image,
      client,
      client_version,
    ]
  }

  depends_on = [
    google_secret_manager_secret_iam_member.run,
    google_secret_manager_secret_version.generated,
    google_project_iam_member.run_sql,
  ]
}

# 来場者のページは Next.js のサーバーから呼ぶ。認証は JWT なので誰でも呼べてよい
resource "google_cloud_run_v2_service_iam_member" "public" {
  name     = google_cloud_run_v2_service.api.name
  location = var.region
  role     = "roles/run.invoker"
  member   = "allUsers"
}

# api.kctfest.jp。先に gcloud domains verify kctfest.jp でドメインの持ち主の確認が要る
resource "google_cloud_run_domain_mapping" "api" {
  name     = "api.${var.domain}"
  location = var.region
  metadata {
    namespace = var.project_id
  }
  spec {
    route_name = google_cloud_run_v2_service.api.name
  }
}
