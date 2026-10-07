data "google_project" "this" {}

resource "google_billing_budget" "monthly" {
  billing_account = var.billing_account_id
  display_name    = "kctfestnav-monthly"
  budget_filter {
    projects = ["projects/${data.google_project.this.number}"]
  }
  amount {
    specified_amount {
      currency_code = var.budget_currency
      units         = tostring(var.budget_amount)
    }
  }
  # 請求先アカウントの管理者にメールが届く
  threshold_rules { threshold_percent = 0.5 }
  threshold_rules { threshold_percent = 0.9 }
  threshold_rules { threshold_percent = 1.0 }
  depends_on = [google_project_service.apis]
}
