# cors-anywhere-go

## Example
 

```
Request examples:

* `http://localhost:7001/http://google.com/` - Google.com with CORS headers
* `http://localhost:7001/google.com` - Same as previous.
* `http://localhost:7001/google.com:443` - Proxies `https://google.com/` 

```

## Start
### Step1 
Install Go 1.16.4

### Step2
To Run this locally simply run:
```
./install.sh
./dev.sh
```

## build && deploy
```
./install.sh
go build -o cors-anywhere-go
./cors-anywhere-go
```


