provider "google" {
  project = var.project_id
  region  = var.region
  # 予算(Billing Budget)の API を自分のアカウントで呼ぶのに要る
  user_project_override = true
  billing_project       = var.project_id
}

# トークンは環境変数 CLOUDFLARE_API_TOKEN で渡す(ファイルに書かない)
provider "cloudflare" {}
