locals {
  # Terraform が作る値
  generated_secrets = {
    DB_PASS          = random_password.db.result
    JWT_SECRET       = random_password.jwt.result
    INTERNAL_API_KEY = random_password.internal_api_key.result
    VOTER_SECRET     = random_password.voter.result
  }
  # 本人が gcloud で手で入れる値(Terraform に書かない)
  # R2_* は Cloudflare の画面で作る R2 の API トークン(写真の置き場所の読み書き)
  manual_secrets = ["REDIS_PASSWORD", "INIT_ADMIN_PASSWORD", "R2_ACCESS_KEY_ID", "R2_SECRET_ACCESS_KEY"]
  all_secrets    = concat(keys(local.generated_secrets), local.manual_secrets)
}

# いいね: Next.js のサーバーだけが知る合言葉(Vercel にも手で入れる)と、投票者番号の署名
resource "random_password" "internal_api_key" {
  length  = 64
  special = false
}

resource "random_password" "voter" {
  length  = 64
  special = false
}

resource "random_password" "jwt" {
  length  = 64
  special = false
}

resource "google_secret_manager_secret" "s" {
  for_each  = toset(local.all_secrets)
  secret_id = lower(replace(each.key, "_", "-"))
  replication {
    auto {}
  }
  depends_on = [google_project_service.apis]
}

resource "google_secret_manager_secret_version" "generated" {
  for_each    = local.generated_secrets
  secret      = google_secret_manager_secret.s[each.key].id
  secret_data = each.value
}

resource "google_secret_manager_secret_iam_member" "run" {
  for_each  = google_secret_manager_secret.s
  secret_id = each.value.id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.run.email}"
}
