# gcvis

Visualise Go program gctrace data in real time

Note: GC timing graphs are only supported for go 1.6

## Usage

Running it directly:

```bash
env GOMAXPROCS=4 gcvis godoc -index -http=:6060
```

Adding the `gctrace` flag yourself:

```bash
GODEBUG=gctrace=1 godoc -index -http=:6060 2>&1 | gcvis
```

Or from a log file:

```bash
GODEBUG=gctrace=1 godoc -index -http=:6060 2> stderr.log
cat stderr.log | gcvis
```

Starting the server without automatically opening a browser:

```bash
gcvis -o=false godoc -index -http=:6060
```
