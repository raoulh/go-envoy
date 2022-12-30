# go-envoy
Pull data from an Enphase IQ Envoy or IQ Combiner. It also provide a simple web server with an API to let 
you use the data from any home automation software of scripts.

The main issue was to query the Enphase Envoy using their new API. The direct polling of data from the local
network was shut off with one of the latest 2022 fw update. From now on, you need to do multiple queries to their
cloud first to get some token. Once you have a valid token you can query the local envoy for data.
This new way of doing is really cumbersome for home automation system, as they do not have the hability to
do so complicated queries to get data.

This project contains a simple CLI tool and a daemon. The CLI tool can query the envoy and then caches the token
to be able to use it the next time. The web daemon should run as a system service, it caches the data and polls
itself for the token/data automatically. You can then use any basic tool to query the endpoints and get data.

Based on https://github.com/cloudkucooland/go-envoy
Heavily modified.

# Usage

First you need to set your credential and envoy serial number.

```
> envoy config set -h=192.168.0.134 -u=xxxx@email.com -s=1234567890 -p=my_super_password
```

Then use any of the CLI to query. The CLI tool can print as raw json too.

```
Usage: envoy [-v] COMMAND [arg...]

Envoy CLI App
                  
Options:          
  -v, --verbose   Verbose debug mode
                  
Commands:         
  config          manage account
  now             display current production
  today           display stats for today production
  info            display info about gateway
  production      display raw json production
  inventory       display raw json inventory
  inverters       display raw json inverters
  home            display raw json /home.json
                  
Run 'envoy COMMAND --help' for more information on a command.
```

```
> envoy now  
🔌Production: 59.43W / 2354W    Consumption: 1689.83W   Net import: 1630.40W
```

## Installation

git clone the repo and type:
```
make
make install
```

## Configuration

Copy `envoy.toml` to `/etc/envoy.toml` and set the correct value in it. The `static` option should be set to
the path where the tool data has been installed (usually in `/usr/local/share/envoy`).

Copy the `envoy.service` file to systemd:
```
cp envoy.service /etc/systemd/system/envoy.service
systemctl daemon-reload
```

# Start
The service can now be enable and started

```
systemctl enable --now envoy.service
```
