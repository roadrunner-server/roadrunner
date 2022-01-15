# Local testing/development files

### Usage:
1. Build the RR binary with the following command: `make build`. As the result you'll see a RR binary.
2. If you are testing you own plugin (Go), you may use `replace` directive in the root `go.mod` file, for example:
```go
module github.com/spiral/roadrunner-binary/v2

go 1.17

require (
	github.com/buger/goterm v1.0.1
	github.com/dustin/go-humanize v1.0.0
	github.com/fatih/color v1.12.0
	github.com/joho/godotenv v1.3.0
	github.com/kami-zh/go-capturer v0.0.0-20171211120116-e492ea43421d
	github.com/mattn/go-runewidth v0.0.13
	github.com/olekukonko/tablewriter v0.0.5
	github.com/spf13/cobra v1.2.1
	// SPIRAL ------------
	github.com/spiral/endure v1.0.3
	github.com/spiral/errors v1.0.12
	github.com/spiral/goridge/v3 v3.2.1
	github.com/spiral/roadrunner/v2 v2.4.0-rc.1
	// ---------------------
	github.com/stretchr/testify v1.7.0
	// SPIRAL --------------
	github.com/temporalio/roadrunner-temporal v1.0.9-beta.1
	// ---------------------
	github.com/vbauerster/mpb/v5 v5.4.0
)

replace github.com/spiral/roadrunner/v2 => ../roadrunner <----- SAMPLE
```

3. Replace sample worker `psr-worker.php` with your application. You can do that by putting all `dev` env into the folder
with your app and replacing lines 21-24 of the `Dockerfile.local` with your app OR by replacing the sample worker. Also, do not forget to update `.rr-docker.yaml`.

5. Next step is to build docker-compose: `docker-compose up`, that's it. After that, you'll have your app running in local dev env with RR.
