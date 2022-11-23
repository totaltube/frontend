Totaltube Frontend
===

The frontend script for Totaltube engine. Hosted on your server and gets information from Totaltube "minion" server 
by API calls. Script holds your site templates in `twig` format, `js`, `css` and all site static files (except thumbs and video).
As a database it uses [badgerdb](https://github.com/dgraph-io/badger). In this database only cache and translations stored.
Typical space needed for the database: from 0.5 to 2 Gb for 100 sites depending on amount of traffic, content and translations. 
Better be sure you have at least 10Gb of free space for this script. 

Installation
___
Just create directory with this structure:
```
root
│   totaltube-frontend
└───sites
│   └───yoursite.com
│       └───extensions
│       └───js
│       └───public
│       └───scss
│       └───templates
└───database
│   config.toml
```
Where `totaltube-frontend` is this script binary (for windows `totaltube-frontend.exe`). 

[Linux binary](https://totaltraffictrader.com/latest/linux/totaltube-frontend.tar.gz) | 
[Windows binary](https://totaltraffictrader.com/latest/windows/totaltube-frontend.zip)

Example site templates can be downloaded [here](https://totaltraffictrader.com/totaltube-download/example-site.zip).
The `config.toml` - is main config file for Totaltube Frontend. It should look like this:
```toml
[general]
port = 8380 # port script will be running on
real_ip_header = "X-Real-Ip" # Header with real IP
nginx = false # totaltube runs under nginx? In dev mode must be false. In production - true, this way script will avoid double redirection if possible by using X-Accel-Redirect header.
use_ipv6_network = true # for IPV6 script will track surfers on one network as same. Better set it true.
api_url = "https://totaltube-test-main.totaltraffictrader.com/api/v1/" # Your totaltube "minion" service API URL
api_secret = "0KzitIKqVkQ28oFwFYzRjzMiqBiKAqRI9U8X57oL" # Your totaltube "minion" service API secret
api_timeout = "30 seconds" # timeout for API response

[frontend]
sites_path = "./sites" # path with sites templates
default_site = "commonsite.dev" # default site
secret_key = "Some secret passthrase - change it to your own" # Your secret key - set it to any
captcha_key = "6LeRDuUhAAAAAK2zqU48jb7db3WXhZBAzcBr2y7Q" # For DMCA, get your own on https://www.google.com/recaptcha (v2)
captcha_secret = "6LeRDuUhAAAAAAsGTo_lq94vadcfclA8l5AEy-Do" # For DMCA, get your own on https://www.google.com/recaptcha (v2)
max_dmca_minute = 5 # All other DMCA's from this IP will be captcha powered.
captcha_whitelist = ["your@email.com"] # Whitelist of emails which doesn't need to be captcha checked.

[database]
path = "./database" # path to database
```
To start Totaltube Frontend, run from this path:
```shell
./totaltube-frontend -c config.toml start
```
Also, you can install Totaltube Frontend as a service on `linux` or `freebsd` by copying `totaltube-frontend` to `/usr/local/bin` and running 
```shell
./totaltube-frontend install
```

[Additional Documentation](docs/Docs.md)