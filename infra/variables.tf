variable "project_id" { type = string }
variable "billing_account_id" { type = string }
variable "region" {
  type    = string
  default = "asia-northeast1"
}
variable "domain" {
  type    = string
  default = "kctfest.jp"
}
variable "cloudflare_account_id" { type = string }
variable "cloudflare_zone_id" { type = string }
variable "github_repo" {
  type    = string
  default = "meg44k/KCTFestNav-Backend"
}
variable "redis_addr" {
  description = "Upstash の host:port"
  type        = string
}
variable "init_admin_id" {
  description = "最初の管理者のログイン ID(パスワードは Secret Manager に手で入れる)"
  type        = string
}
variable "min_instances" {
  description = "当日の 2 日間だけ 1 にする"
  type        = number
  default     = 0
}
variable "max_instances" {
  type    = number
  default = 5
}
variable "budget_amount" {
  type    = number
  default = 3000
}
variable "budget_currency" {
  description = "請求先アカウントの通貨と同じにする(USD の口座なら USD・20 など)"
  type        = string
  default     = "JPY"
}
variable "vercel_apex_ip" {
  description = "Vercel の画面に出る kctfest.jp 用の A レコードの値"
  type        = string
  default     = "76.76.21.21"
}
