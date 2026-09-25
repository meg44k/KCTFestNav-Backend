#!/bin/bash
# db/schema.sql を開発用DBと各テストDBに適用する。
set -euo pipefail

for db in kctfestnav kctfest_test_handler kctfest_test_repository; do
  echo "[init] applying schema.sql to ${db}"
  mysql --protocol=socket --default-character-set=utf8mb4 -u root "${db}" < /etc/kctfestnav/schema.sql
done
