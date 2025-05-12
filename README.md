# demo
Opsmind Demo app

```
docker run --network=host --rm schemathesis/schemathesis:stable \
run -c all --experimental=openapi-3.1 \
--base-url=http://127.0.0.1:8080 https://raw.githubusercontent.com/opsminded/spec/refs/heads/main/openapi.json
```
