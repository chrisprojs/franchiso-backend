Set environment variable before running:
- `DATABASE_URL` example: `postgres://user:password@localhost:5432/nama_db?sslmode=disable`
- `ELASTIC_URL` example: `http://localhost:9200`

Run Application:

```
go run .
```

Run this in terminal:
```
ngrok http http://localhost:8080
```

Add this notification url to midtrans
```
https://880034e3cbd1.ngrok-free.app/mid_trans/call_back
```
