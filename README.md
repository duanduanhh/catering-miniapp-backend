# Start

```makefile
make build
make build-linux
make swag
make docker
```

# Run

```dockerfile
docker run -d --name miniapp-backend-test -p 16789:8000 -e APP_ENV=test --restart=always swr.cn-north-1.myhuaweicloud.com/catering-cyxx/miniapp-backend:v1-20250114
```

