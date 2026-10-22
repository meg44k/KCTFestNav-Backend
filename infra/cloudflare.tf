# Vercel と Google の証明書を通すため、どれも Cloudflare のプロキシを通さない(proxied = false)
resource "cloudflare_dns_record" "apex" {
  zone_id = var.cloudflare_zone_id
  name    = var.domain
  type    = "A"
  content = var.vercel_apex_ip
  ttl     = 1
  proxied = false
}

resource "cloudflare_dns_record" "www" {
  zone_id = var.cloudflare_zone_id
  name    = "www.${var.domain}"
  type    = "CNAME"
  content = var.vercel_www_cname
  ttl     = 1
  proxied = false
}

resource "cloudflare_dns_record" "api" {
  zone_id = var.cloudflare_zone_id
  name    = "api.${var.domain}"
  type    = "CNAME"
  content = "ghs.googlehosted.com"
  ttl     = 1
  proxied = false
}

# 写真の置き場。使い方は写真の設計で決める
resource "cloudflare_r2_bucket" "photos" {
  account_id = var.cloudflare_account_id
  name       = "kctfestnav-photos"
  location   = "apac"
}

# img.kctfes.app(DNS のレコードは Cloudflare が作る)
resource "cloudflare_r2_custom_domain" "img" {
  account_id  = var.cloudflare_account_id
  bucket_name = cloudflare_r2_bucket.photos.name
  domain      = "img.${var.domain}"
  zone_id     = var.cloudflare_zone_id
  enabled     = true
}
