#!/bin/bash

URL="http://localhost:8000/api/goods"

declare -a DATA_ARRAY=(
  '{"match": "Bork", "reward": 10, "reward_type": "%"}'
  '{"match": "Apple", "reward": 50, "reward_type": "pt"}'
  '{"match": "Samsung", "reward": 15, "reward_type": "%"}'
  '{"match": "Huawei", "reward": 100, "reward_type": "pt"}'
)

for DATA in "${DATA_ARRAY[@]}"; do
  echo "Sending request with data: $DATA"
  curl -X POST "$URL" \
    -H "Content-Type: application/json" \
    -d "$DATA"
  echo ""  # Newline for readability
done