output "api_url" { value = google_cloud_run_v2_service.api.uri }
output "run_service" { value = google_cloud_run_v2_service.api.name }
output "sql_connection_name" { value = google_sql_database_instance.main.connection_name }
output "registry" { value = "${var.region}-docker.pkg.dev/${var.project_id}/${google_artifact_registry_repository.backend.repository_id}" }
output "wif_provider" { value = google_iam_workload_identity_pool_provider.github.name }
output "deploy_service_account" { value = google_service_account.deploy.email }
