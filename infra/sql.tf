resource "google_sql_database_instance" "main" {
  name             = "kctfestnav"
  database_version = "MYSQL_8_4"
  region           = var.region
  # 文化祭のあとに terraform destroy で消すため
  deletion_protection = false

  settings {
    # 8.4 の既定は Enterprise Plus(高い)なので明示する
    edition           = "ENTERPRISE"
    tier              = "db-f1-micro"
    availability_type = "ZONAL"
    disk_size         = 10
    disk_type         = "PD_SSD"
    backup_configuration {
      enabled    = true
      start_time = "18:00" # 日本時間 3:00
    }
    # 公開 IP は持つが許可するネットワークは無し。Cloud Run と Auth Proxy は IAM でつなぐ
    ip_configuration {
      ipv4_enabled = true
    }
    database_flags {
      name  = "character_set_server"
      value = "utf8mb4"
    }
  }
  depends_on = [google_project_service.apis]
}

resource "google_sql_database" "app" {
  name      = "kctfestnav"
  instance  = google_sql_database_instance.main.name
  charset   = "utf8mb4"
  collation = "utf8mb4_0900_ai_ci"
}

resource "random_password" "db" {
  length  = 32
  special = false
}

resource "google_sql_user" "app" {
  name     = "kctfestnav"
  instance = google_sql_database_instance.main.name
  password = random_password.db.result
}
