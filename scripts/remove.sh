curl -X DELETE http://0.0.0.0:8080/remove \
    -H "Content-Type: application/json" \
    -d '{"service_name": "test_service"}' \
    -i