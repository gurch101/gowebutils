#!/bin/bash

seed_db() {
  data_dir="./db/data"

  if [[ ! -d "$data_dir" ]]; then
      echo "Data directory not found: $data_dir"
      exit 1
  fi


  for file in "$data_dir"/*.sql; do
      if [[ -f "$file" ]]; then
          echo "Executing SQL file: $file"
          if ! sqlite3 "$db_name" < "$file"; then
              echo "Failed to execute SQL file: $file"
              exit 1
          fi
      fi
  done
}

if [[ "$#" -ne 1 ]]; then
    echo "Usage: $0 <db_name>"
    exit 1
fi

db_name="$1"

seed_db "$db_name"
