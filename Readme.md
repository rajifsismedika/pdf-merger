### Run

go run cmd/server/main.go

Edit the .env

```.env
PORT=8080
BASE_URL=https://devone.aplikasi.web.id

```

Adjust accordingly.

### Test using curl  

Create curl-format.txt file

```txt
   time_namelookup:  %{time_namelookup}s\n
        time_connect:  %{time_connect}s\n
     time_appconnect:  %{time_appconnect}s\n
    time_pretransfer:  %{time_pretransfer}s\n
       time_redirect:  %{time_redirect}s\n
  time_starttransfer:  %{time_starttransfer}s\n
                     ----------\n
          time_total:  %{time_total}s\n
```

```bash
curl -w "@curl-format.txt" -o /dev/null http://localhost:8080/report/1
curl -w "@curl-format.txt" -o /dev/null http://localhost:8080/report/1000
```

dimana :
1 dan 1000 adalah mergeRequestID dari database one yang aktif

### Build

static

```bash
CGO_ENABLED=0 go build -o ./build/api_merge_server cmd/server/main.go
```

CGO_ENABLED=0 go build artinya menonaktifkan penggunaan Cgo saat build binary Go.
- static binary (tanpa libc, cocok untuk Docker, Alpine)
- Cross compilation ke OS/arch lain
- Portabilitas lebih tinggi