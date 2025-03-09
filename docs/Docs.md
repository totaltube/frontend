# Totaltube Frontend Documentation

Glossary
* [Server Configuration and Deployment](#server-configuration-and-deployment)
* [Global Configuration](#global-configuration)
* [Command Line Interface](#command-line-interface)
* [HTTPS Configuration](#https-configuration)
* [Site configuration](#site-configuration)
* [Site templates](#site-templates)
* [Available special tags in templates](#available-special-tags-in-templates)
* [Functions, available in templates and custom functions.](#functions-available-in-templates-and-custom-functions)
* [Variables, available in template files and custom functions.](#variables-available-in-template-files-and-custom-functions)
* [Special variables, available in different template files.](#special-variables-available-in-different-template-files)
* [Custom functions](#custom-functions)
* [Types, used in site templates and custom functions](Types.md)
* [Site javascript build system](#site-javascript-build-system)
* [Site css build system](#site-css-build-system)

## Server Configuration and Deployment

### Running in Development Mode

To start Totaltube Frontend in development mode, run:

```shell
./bin/totaltube-frontend -c global-config.toml start
```
In development mode, compilation errors for JavaScript and CSS will be output to the console window.

### Running as a Service
To install Totaltube Frontend as a service on Linux or FreeBSD:
1. Copy the binary to /usr/local/bin:
```cp bin/totaltube-frontend /usr/local/bin/```
2. Install as a service:
```totaltube-frontend install```
3. Start the service:
```systemctl start totaltube-frontend```
4. View logs:
```journalctl -u totaltube-frontend -f```

## Global Configuration

The global configuration file (`global-config.toml`) contains settings for the entire application:

```toml
[general]
port = 8380 # HTTP server port
real_ip_header = "" # Header to get real client IP
api_url = "http://minion-api-server/api/v1" # URL of Minion API
api_secret = "secret" # Secret key for Minion API
api_timeout = "5s" # API request timeout
debug = false # Enable debug mode

[frontend]
sites_path = "sites" # Path to sites directory
default_site = "example.com" # Default site
secret_key = "random-secret" # Secret key for security
captcha_key = "" # reCAPTCHA site key
captcha_secret = "" # reCAPTCHA secret key
max_dmca_minute = 5 # DMCA requests limit per minute
captcha_whitelist = [] # Email whitelist for DMCA

[database]
path = "database" # Database files path
backup_path = "database-backup" # Backup path
```

## Command Line Interface
Totaltube Frontend supports the following commands:
```
start       - start the server
stop        - stop the server
install     - install as a service
uninstall   - remove service
version     - show version information
help        - show help information
```
Options:
```
-c, --config FILE   - path to config file (default: global-config.toml)
-d, --debug         - enable debug mode
```
## HTTPS Configuration

For production environments, it's recommended to run Totaltube Frontend behind a reverse proxy like Nginx that handles HTTPS:

```nginx
server {
    listen 443 ssl;
    server_name your-site.com;

    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;

    location / {
        proxy_pass http://localhost:8380;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```
## Site configuration

In site templates directory must be `config.toml`. In this file in `[routes]` section you should define your routes for different pages of your site. Example:
```toml
[routes]
# if url is "" - this means it will not be served. Also you can just not specify url for some route to not serve it.
top_categories = "/" # URL of page with top categories sorted by ctr
top_content = "/top" # URL of page with top content sorted by ctr.
autocomplete = "/autocomplete" # URL of autocomplete api
search = "/search/{query}" # URL of search page
popular = "/best" # URL of popular page
new = "/new" # URL of new content page
long = "/long" # URL of long content page
model = "/pornstar/{slug}" # URL of model page, must have {slug} or {id} param
models = "/porstar-list" # URL of models page
category = "/category/{slug}" # URL of category page, must have {slug} or {id} param
channel = "/channel/{slug}" # URL of channel page, must have {slug} or {id} param
content_item = "/content/{category}/{slug}" # URL of content item page, must have {slug} or {id} param. Can have {category} param - main category of content, optional
fake_player = "/player/{slug}" # URL of fake video player, can have {slug} or {id} param
video_embed = "/embed/{slug}" # URL of video embed for hosted video
dmca = "/dmca" # dmca report uri
out = "/c" # URL of out script
language_template = "/{lang}{route}" #template for language id in route for multilingual sites if multi_language is true
```
If route is blank - it will not be served by script.

In `[routes.custom]` section you can define some custom routes in format `route_name = "/path/to/site"`. For each of these routes you must create
file `route-{route_name}.js` in which must be 4 functions:
- `cacheKey()` must return the key for caching page on this route. Use route params or querystring params to create some unique cache key.
- `cacheTtl()` must return the ttl for cache of page.
- `prepare()` must return object with some custom vars available in page template.
- `render()` defines how to render this page. If this function returns nothing, script will try to render template with name `custom-{route_name}.twig`.
  If this function returns string or some other data - it will be used as output. If this function returns object, json will be the output.

In `[params]` section you can override default querystring param names for some pages. The section should look like this:
```toml
[params] # Query params name override
content_slug = "slug"
category_slug = "category"
model_slug = "model"
model_id = "model_id"
channel_slug = "channel"
channel_id = "channel_id"
duration_gte = "duration_from"
duration_lt = "duration_to"
search_query = "q"
sort_by = "sort"
sort_by_views = "views"
sort_by_views_timeframe = "timeframe"
sort_by_duration = "duration"
sort_by_date = "date"
sort_by_rand = "rand"
page = "page"
nocache = "nocache"
# If you will modify below values you must modify them also in default main.ts
content_id = "id"
category_id = "cid"
count_redirect = "r" # param of out script with encoded redirect url
count_type = "t" # param of out script with type of click counting
count_type_category = "c" # content on category page click type
count_type_top_categories = "tca" # category on top categories page click type
count_type_top_content = "tc" # content on top content page click type
count_thumb_id = "tid" # thumb id of clicked content link
count_trade = "tr"  # redirect to trade script
```

In `[general]` section you can define some general options like this:
```toml
[general]
trade_url_template = "/ttt/o?url={{encoded_url}}" # trade script url template,
# can have {{encoded_url}} for url encoded redirect url or {{url}} as raw redirect url
multi_language = true # true if this site is multilingual
minify_html = false # if yes, html output from template will be minified (makes a little impact on server cpu usage)
pagination_max_rendered_links = 15 # Maximum rendered page links, default 10
models_per_page = 20 # Number of models per page on models page
content_related_amount = 20 # Number of related videos on content item page
fake_video_page = true # Show fake video page for video links
disable_categories_redirect = false # if true - redirect to category from top categories page based on referrer will be disabled.
api_url = "" # if set, it will override minion api url in global config
api_secret = "" # if set, it will override minion api secret in global config
```

In `[javascript]` and `[css]` sections you can define options to build js and css for the site. Example:
```toml
[javascript]
entries = ["main.ts", "video.ts", "video-hosted.ts", "gallery.ts"] # main javascrypt/typescrypt source entry files to build the bundle/bundles
destination = "" # destination path relative to public path. If empty - destination path = public path.
minify = true # minify resulting javascript

[scss]
entries = ["main.scss"] # scss entries to build css
destination = "" # destination path relative to public. If empty - place result css in public with same name as entry
images_path = "images" # images resolution from this path relative to scss path
fonts_path = "fonts" # fonts resolution from this path relative to scss path
minify = true # minify resulting css
```
In `[sitemap]` section you can define some options for building the sitemap.xml. Example:
```toml
[sitemap]
route = "/sitemap.xml" # route for main sitemap.xml file, if empty - sitemap will not be generated. Default - /sitemap.xml
additional_links = ["/somelink", "/someotherlink"] # additional links to add to sitemap. By default, sitemap will be generated for main links, top categories, top models, top channels, top searches and last videos.
max_links = 100 # max links in one sitemap file
categories_amount = 100 # num top categories links to place in sitemap
models_amount = 100 # num top models links to place in sitemap
channels_amount = 100 # num top channels links to place in sitemap
searches_amount = 100 # num top searches links to place in sitemap
last_videos_amount = 500 # num last videos links to place in sitemap
# for all these categories_amount, models_amount, etc.. if you set to 0, these links will not be added to sitemap.
```
In `[custom]` section you can define some custom options, which will be available in templates as config.Custom.your_option. Example:
```toml
[custom] # Some custom configuration options available in template as config.Custom.your_option
site_title = "Common site"
ttt_in_uri = "/ttt/in"
thumb_rotate_delay = "1500"
ttt_secret = "some secret"
ttt_tds_uri = "/ttt/tds"

vast_url = "https://syndication.realsrv.com/splash.php?idzone=4232570"
cdn_salt = "JsjdIyu872@jkshHHsl;"
default_video_format = "main" # default format for hosted video.
default_gallery_format = "main" #default format for picture gallery
```

## Site templates

In templates path you can define site templates with [django](https://django.readthedocs.io/en/1.7.x/topics/templates.html)-like syntax. Actually [pongo2](https://github.com/flosch/pongo2) go library is used. 
Template names:
* `404.twig` - for 404 errors
* `500.twig` - for server errors
* `category.twig` - for category page
* `channel.twig` - for channel page
* `content-item` - for page showing content item (video or gallery)
* `fake-player.twig` - if `fake_video_page` config option is true this template will serve fake video page for content of type `video-link`, which should show fake video with link to actual video on other site.
* `long.twig` - for video sorted by duration
* `model.twig` - for model page (content belongs to particular model)
* `models.twig` - for models listing page
* `new.twig` - for content sorted by date
* `popular.twig` - for content sorted by views
* `search.twig` - for content containing some query
* `top-categories.twig` - for top categories page (categories sorted by CTR)
* `top-content.twig` - for top content page (content sorted by CTR)
* `video-embed` - for video embed page for hosted video
For custom routes you can create `custom-{route_name}.twig` files.
We recommend to place some other template files, which can contains some common macros, in separate directory `common`. You can include templates and macros from this directory.

## Available special tags in templates

Among standard [django template tags](https://django.readthedocs.io/en/1.7.x/topics/templates.html#tags) Totaltube frontend templates can have special tags: 

#### `{% fetch %}...{% endfetch %}`
this tag can be used to fetch some data from Totaltube "minion" service or from any other server in Internet. First argument of this tag is site URL from where to fetch data or special words to fetch some data from Totaltube "minion". After this string argument can be added comma separated additional named params in format `name = val`. Common params for fetching from URL:
* `cache` - set cache timeout in seconds. If not set, no caching of this request will be performed, but additional caching of full template output can take place.
* `header` - set http request header in format `Header-Name: header_value`. You can add several `header` params to add several headers. This param is for fetching from URL only.
* `timeout` - set http request timeout in human-readable format, like "30 seconds"
* `method` - set http method if you fetch from URL.
* `raw` - is set to true, response from URL will be raw string, no JSON unmarshalling. By default, if response is JSON, `fetch_response` variable inside `fetch` tag will contain object with JSON data.
* all other unknown params will be treated as querystring params appended to URL for `GET` and `DELETE` requests and as JSON or form data for `POST` and `PUT` requests. If you need to send form data, add `header` param with value `"Content-Type: application/x-www-form-urlencoded;charset=UTF-8"`. If you need to add querystring or JSON/form params with same names as mentioned params, just add `arg_` prefix to these params (i.e. `arg_cache = "some-value"`)

Example of `fetch` tag with fetching some data from URL:
```django
{# for example, this API returns JSON with origin param set to requester IP address #}
{% fetch "http://httpbin.org/ip", header = "Accept: application/json", timeout = "5 seconds", cache = 60 %}
    {# the response data will be inside var fetch_response #}
    {{ fetch_response.origin }} {# here template will output your server IP address #}
{% endfetch %}
```

First argument can be not URL, but special word to fetch some data from Totaltube "minion". These words are: 
* `"content"`: fetch some content from minion. Available params:
  * `cache` - cache timeout in seconds. 
  * `category_slug` - slug of category to which content belongs
  * `category_id` - numeric id of category to which content belongs
  * `channel_slug` - slug of content channel
  * `channel_id` - numeric id of content channel
  * `model_slug` - slug of model which appears on this content
  * `model_id` - numeric id of model which appears on this content
  * `tag` - filter content by this tag
  * `duration_gte` - filter content by duration greater or equal than this (in seconds)
  * `duration_lt` - filter content by duration less than this (in seconds)
  * `amount` - amount of content to fetch (on one page), default 100.
  * `page` - page number to fetch from 1. Default 1.
  * `lang` - language to fetch content for. Default - current requested page language or "en".
  * `sort` - sort order. Can be `"popular"` (by ctrs desc), `"views"` (by views desc), `"dated"` (by content date desc), `"duration"` (by duration desc), `"rand"` (random), `"rand1"` (random without supporting pagination - more fast).
  * `timeframe` - timeframe для сортировки по `"views"` ("all", "hour" или "month"). По умолчанию "month". 
  
  result content will be in variable `fetched_content`. The type of variable is [ContentResults](Types.md#contentresults). Example of `fetch` tag to fetch content:
  ```django
  {% fetch "content", sort = "rand1", cache = 30, amount = 20 %}
  {% for item in fetched_content.Items %}
    ...
  {% endfor %}
  {% endfetch %}
  ```
* `"categories"`: fetch categories from minion API. Available params:
  * `cache` - cache timeout in seconds.
  * `lang` - language to fetch categories for. Default - current requested page language or "en".
  * `amount` - amount of categories to fetch (on one page), default 100.
  * `page` - page number to fetch from 1. Default 1.
  * `sort` - sort order. Can be `"popular"` (by ctrs desc), `"title"` (by title asc), `"total"` (by total content in category desc)
  
  result data will be in variable `categories`. The type of variable is [CategoryResults](Types.md#categoryresults). Example of `fetch` tag to fetch categories: 
  ```django
  {% fetch "categories", sort = "popular", cache = 300, amount = 200 %}
  {% sort categories.Items, type = "title" %}
  {% for item in categories.Items %}
    ...
  {% endfor %}
  {% endfetch %}
  ```
* `"models"`: fetch models from minion API. Available params:
  * `cache` - cache timeout in seconds. 
  * `lang` - language to fetch models for. Default - current requested page language or "en".
  * `amount` - amount of models to fetch (on one page), default 100.
  * `page` - page number to fetch from 1. Default 1.
  * `sort` - sort order. Can be `"popular"` (by ctrs desc), `"title"` (by title asc), `"total"` (by total content with this model)

  result data will be in variable `models`. The type of variable is [ModelResults](Types.md#modelresults). Example of `fetch` tag to fetch models:
  ```django
  {% fetch "models", sort = "popular", cache = 300, amount = 200 %}
  {% sort models.Items, type = "title" %}
  {% for item in models.Items %}
    ...
  {% endfor %}
  {% endfetch %}
  ```
* `"channels"`: fetch channels from minion API. Available params:
  * `cache` - cache timeout in seconds.
  * `lang` - language to fetch channels for. Default - current requested page language or "en".
  * `amount` - amount of channels to fetch (on one page), default 100.
  * `page` - page number to fetch from 1. Default 1.
  * `sort` - sort order. Can be `"popular"` (by views desc), `"title"` (by title asc), `"total"` (by total content in this channel) 

  result data will be in variable `channels`. The type of variable is [ChannelResults](Types.md#channelresults). Example of `fetch` tag to fetch channels:
  ```django
  {% fetch "channels", sort = "popular", cache = 300, amount = 50 %}
  {% sort channels.Items, type = "title" %}
  {% for item in channels.Items %}
    ...
  {% endfor %}
  {% endfetch %}
  ```
* `"searches"`: fetch search queries of surfers on your site. Available params:
  * `cache` - cache timeout in seconds.
  * `lang` - language to fetch search queries for. Default - current requested page language or "en".
  * `amount` - max amount of search queries to fetch. Default 100.
  * `sort` - sort order. Can be `"popular"` (by number of searches with same query) and `"random"`
  * `min_searches` - filter results by minimum number of searches of same query. Default - 1.

  result data will be in variable `searches`. This is array of [TopSearch](Types.md#topsearch) objects.
Example of `fetch` tag to fetch searches:
  ```django
  {% fetch "searches", amount = 40, page = 1, sort = "rand", cache = 300, min_searches = 1 %}
    {# inside this tag they will be available with variable name searches #}
    <h3 class="list-title"><i class="icon-search padded"></i>{{ translate('Trends') }}</h3>
      <ul class="list">
      {% for item in searches %}
        <li>
          <a href="{% link "search", query = item.Message %}">{{ item.Message | capfirst }}</a>
            <span class="total" data-number="{{ item.Searches }}">{{ item.Searches }}</span>
        </li>
      {% endfor %}
      </ul>
  {% endfetch %}
  ```

#### `{% link %}`
this tag is used to generate link to some site page, based on route settings in `config.toml` of the site. It just outputs this link, or can save it to specified variable (param `as`). First argument of this tag is the name of the route, which defined in `[routes]` section of `config.toml`, or `"custom.{custom_route_name}"` for custom route, or `"current"` for current page route. Other possible named parameters:
* `as` - do not write the link, but save it to variable instead. Variable name is the value of this param.
* `out` - boolean param. If true, link will go via out script to count the click to content or category (for CTR counting). Default false.
* `with_trade` - boolean param. If true, link will go via trade script.
* `full_url` - boolean param. If true, link will be generated as full url, with https://your-domain/ prefix. Default false.
* All other params will be treated as route named params or additional querystring params. If querystring params are with same names as in `[params]` section of `config.toml`, they will be replaced with corresponding values. 

Examples of `link` tag:
```django
<a href="{% link "search", query = "some query" %}">Link to search page to find some query</a>
<a href="{% link "model", slug = "eva" %}">Link to model with slug eva</a>
<a href="{% link "category", slug = "red", out = true, with_trade = true %}">Link to category red with click count and via trade script</a>
```

#### `{% sort %}`
this tag is used to sort arrays of [categories](Types.md#categoryresult), [models](Types.md#modelresult) and [channels](Types.md#channelresult) by title or total content. First argument is what to sort, and another one can be `type` - type of sort - `"title"` or `"total"`. Example:
```django
{% fetch "categories", amount = 200, page = 1, sort = "popular", cache = 300 %}
{% sort categories.Items, type = "title" %}
{# now categories.Items sorted by title #}
...
{% endfetch %}
```

#### `{% alternate %}`
this tag outputs alternate page link for given language for current page (for multilingual sites). Search page doesn't have alternate. Language id is the only argument of this tag. Example:
```django
{% for lang in languages %}
  <li>
    <a hreflang="{{ lang.Id }}" href="{% alternate lang.Id %}"
      data-lang="{{ lang.Id }}">
      <img src="{{ static("flags/", lang.Country, ".png") }}"
        loading="lazy">
      {{ lang.Native | capfirst }}
    </a>
  </li>
{% endfor %}
```

#### `{% alternates %}`
this tag outputs `<link rel="alternate">` tags for all languages for current page. Useful to put it in `<head>` section of your site. It has no params. Example:
```django
<head>
{% alternates %}
...
</head>
```

#### `{% canonical %}`
this tag outputs `<link rel="canonical">` tag for current page. Useful to put it in `<head>` section of your site. It has no params. Example:
```django
<head>
{% canonical %}
...
</head>
```

#### `{% prevnext %}`
this tag outputs `<link rel="prev">` and `<link rel="next">` tags for current page if it has pagination. Useful to put it in `<head>` section of your site. It has no params. Example:
```django
<head>
{% prevnext %}
...
</head>
```

#### `{% page_link %}`
this tag outputs link to current route with another page parameter for pagination. The only argument is the page number. Example:
```django
 <a href="{% page_link 2 %}">Link to page 2</a>
```

#### `{% repeat %}`
this tag generates array with same value with given size and saves it as variable of given name. Example:
```django
{% repeat "test", amount = 100, as = "test_array" %}
{% for item in test_array %}
{{ item }}
{% endfor %}
{# this will output word "test" 100 times #}
```

#### `{% dilute %}`
this tag is used to dilute some elements from one array to another array in random order. It is useful if you want, for example, output content thumbs, but to show between them toplist banners or native ads in random order. First argument is what array to dilute (original), second is the array with which to dilute (donor). Other params:
* `as` - name of variable where to save diluted array. Required.
* `from` - index of original array to start dilute from (zero based, including). Default: 0. It is the index from which in result array can appear elements from donor array.
* `to` - index of original array to end dilute to (zero based, not including). Default: length of original array.
* `max` - maximum elements from donor to dilute.

Example of using `dilute` tag:
```django
{# this code outputs thumbs with some ads and toplist items between them in random order #}
{% set toplist = fetch_toplist("", 8, true) %}
{% repeat "ad", amount = 4, as = ads_list %}
{% dilute content.Items, ads_list, from = 0, to = 50, as = items %}
{% dilute items, toplist.items, from = 0, to = 30, as = items %}
{% for item in items %}
    {% if item == "ad" %}
        {{ show_ad("native1", forloop.Counter, false) }}
    {% elif item.trader.domain %}
        {{ show_toplist_thumb(item, toplist, false) }}
    {% else %}
        {{ show_content(item) }}
    {% endif %}
{% endfor %}
```

#### `{% dynamic %}`
this tag is used to bypass cache for some expression. The arguments of this tag is any expression which can be used inside `{{ }}` brackets in django template. Usually it should be some custom function or macro. This expression will be evaluated on each request, without cache. Example:
```django
{# this will insert Total Traffic Trader js code and count In. insert_ttt is defined in extensions/function-insert_ttt.js of example site template #}
{% dynamic insert_ttt() | safe %} 
```
Also, special case is to use `include` keyword, followed by string with template name to include. In this case, this template name will be evaluated on each request (no cache) with any function calls inside. Template must be in `templates` path, no sub-paths. Example:
```django
{# this will insert contents of template custom-dynamic-insert.twig evaluated on each request #}
{% dynamic include "custom-dynamic-insert.twig" %}
```

## Functions, available in templates and custom functions.

Besides standard django functions, templates can use some additional:
* `flate` - flate compression of raw data in bytes. Result is bytes.
* `deflate` - flate decompression of raw data in bytes. Result is bytes.
* `gzip` - gzip compression of raw data in bytes. Result is bytes.
* `ungzip` - gzip decompression of raw data in bytes. Result is bytes.
* `zip` - zip compression of raw data in bytes. Result is bytes.
* `unzip` - zip decompression of raw data in bytes. Result is bytes.
* `base64` - base64 compression of raw data in bytes. Result is string.
* `sha1` - sha1 hash of string. Result is hex-encoded string.
* `sha1_raw` - sha1 hash of string. Result is raw bytes.
* `md5` - md5 hash of string. Result is hex-encoded string.
* `md5_raw` - md5 hash of string. Result is raw bytes.
* `md4` - md4 hash of string. Result is hex-encoded string.
* `md4_raw` - md4 hash of string. Result is raw bytes.
* `sha256` - sha256 hash of string. Result is hex-encoded string.
* `sha256_raw` - sha256 hash of string. Result is raw bytes.
* `sha512` - sha512 hash of string. Result is hex-encoded string.
* `sha512_raw` - sha512 hash of string. Result is raw bytes.
* `time8601` - takes time as parameter and returns time, formatted with 8601 standard (`"2006-01-02T15:04:05"`)
* `duration8601` - takes duration as seconds or ContentDuration and returns duration formatted with 8601 standard.
* `translate` - takes string as text to translate and translates it to the language of current request. The translation type is `page-text`. 
It's deferred translate, so if the text is not translated yet, it will show untranslated text and will queue text translation.
* `translate_title` - same as `translate`, but translation type is `content-title`.
* `translate_description` - same as `translate`, but translation type is `content-description`.
* `translate_query` - same as `translate`, but translation type is `query` (for translating search queries).
* `static` - makes link to static content with additional parameter, preventing caching of changed static file. Example:
```django
<link type="text/css" rel="stylesheet" href="{{ static("main.css") }}">
```
* `len` - counts len of array
* `time_ago` - takes time as input and outputs date in format time ago (i.e. "5 days ago") localized to current request language.
* `pagination` - outputs array of pagination links info for current route. Each element of this array contains params Type ("prev" for prev link, "next" for next link, "ellipsis" for ellipsis, "page" for page link), State ("default" or "disabled"), Page (page number). Example of using `pagination` function:
```django
<ul class="pagination">
    {% for item in pagination() %}
        {% if item.Type == "prev" %}
            <li class="{{ item.State }}">
                {% if item.State == "disabled" %}
                    <i class="icon-circle-left"></i>
                {% else %}
                    <a href="{% page_link item.Page %}"><i class="icon-circle-left"></i></a>
                {% endif %}
            </li>
        {% endif %}
        {% if item.Type == "next" %}
            <li{% if item.State != "default" %} class="{{ item.State }}"{% endif %}>
                {% if item.State == "disabled" %}
                    <i class="icon-circle-right"></i>
                {% else %}
                    <a href="{% page_link item.Page %}"><i class="icon-circle-right"></i></a>
                {% endif %}
            </li>
        {% endif %}
        {% if item.Type == "ellipsis" %}
            <li>
                ...
            </li>
        {% endif %}
        {% if item.Type == "page" %}
            <li{% if item.State != "default" %} class="{{ item.State }}"{% endif %}>
                <a href="{% page_link item.Page %}" data-number="{{ item.Page }}">{{ item.Page }}</a>
            </li>
        {% endif %}
    {% endfor %}
</ul>
```
* `parse_iframe` - parses code with iframe and returns object with fields Src (iframe src), Width (iframe width) and Height (iframe height)
* `set_cookie` - function will set cookie. First param is the name of cookie, second - value, third - expire as time or duration. This function is useful only in dynamic context (no cache), so it must be used in macro or in custom function in conclusion with `{% dynamic %}` tag.
* `set_var` - takes variable name and value and saves it in current context, so it can be used later in another custom function or in template. Example: `set_var("var_name", "var_value")`
* `get_var` - returns value of variable, saved with `set_var` function. Example: `{{ get_var("var_name") }}`
* `get_content` - function to get content. It accepts variable amount of params as pairs where first is param name and second is param value. Possible param names are
`"lang"`, `"page"`, `"amount"`, `"category_id"`, `"category_slug"`, `"channel_id"`, `"model_id"`, `"channel_slug"`, `"model_slug"`, `"related_message"`, `"sort"`, `"timeframe"`, `"tag"`, `"duration_gte"`, `"duration_lt"`, `"search_query"`. The meaning of params is the same as in `{% fetch "content" %}` tag. Example: 
```django
{# get some random content #}
{% set content = get_content("sort", "rand1", "amount", 10) %}
{% for item in content.Items %}
...
{% endfor %}
```
* `get_top_content` - function to get top content sorted by CTR. It accepts variable amount of params as pairs where first is param name and second is param value. Possible param names are `"lang"` and `"page"`. The meaning of params is the same as in `{% fetch "content" %}` tag. Result is of type [ContentResults](Types.md#contentresults).
* `get_top_categories` - function to get top categories sorted by CTR. It accepts variable amount of params as pairs where first is param name and second is param value. Possible param names are `"lang"` and `"page"`. The meaning of params is the same as in `{% fetch "categories", sort = "popular" %}`. Result is of type [CategoryResults](Types.md#categoryresults)
* `get_content_item` - function to get content item information from Totaltube "minion" API. It accepts variable amount of params as pairs where first is param name and second is param value. Possible param names are:
  * `"lang"` - language of content item, defaults to current page language.
  * `"id"` - id of content (either `id` or `slug` is required)
  * `"slug"` - slug of content (either `id` or `slug` is required)
  * `"related_amount"` - amount of related content to get
  * `"orfl"` - means "Omit related for link ". If set to true and content is of type `"video-link"`, the related content will not be fetched.
  The result is of type [ContentItemResult](Types.md#contentitemresult)
* `get_models_list` - function to get models list. It accepts variable amount of params as pairs where first is param name and second is param value. Possible param names are `"lang"`, `"page"`, `"amount"`, `"sort"` (can be `"title"`, `"total"`, `"popular"`), `"search_query"` (to search model by chars in the name). The result is of type [ModelResults](Types.md#modelresults)
* `get_categories_list` - function to get categories list. It accepts variable amount of params as pairs where first is param name and second is param value. Possible param names are `"lang"`, `"page"`, `"amount"`, `"sort"` (can be `"title"`, `"total"`, `"popular"`). The result is of type [CategoryResults](Types.md#categoryresults).
* `get_channels_list` - function to get channels list. It accepts variable amount of params as pairs where first is param name and second is param value. Possible param names are `"lang"`, `"page"`, `"amount"`, `"sort"` (can be `"title"`, `"total"`, `"popular"`). The result is of type [ChannelResults](Types.md#channelresults)
* `get_category_top` - function to get top content in category. It accepts variable amount of params as pairs where first is param name and second is param value. Possible param names are:
  * `lang` - language of content items, defaults to current page language.
  * `page` - for pagination from 1, default 1.
  * `category_id` - category numeric id for which to get the top. Either `category_id` or `category_slug` is required.
  * `category_slug` - category slug for which to get the top. Either `category_id` or `category_slug` is required.
  The result is of type [ContentResults](Types.md#contentresults)
* `add_random_content` function to add random content to fetched content items. First argument is array of [content items](Types.md#contentresult) and the second is amount of items required in final result. Second argument can be omitted to use default amount for category layout. Result is array of [ContentResult](Types.md#contentresult).
* `merge` - function to merge two arrays into one by appending second array to the first. The result is the merged array.
* `link` - function to get URL to some site page or to any external page with passed params. Same as [`{% link %}`](#-link-) tag. First argument is the route name or any external URL. All other parameters - is pairs of key/value for route params and querystring params. Absolutely the same as with [`{% link %}`](#-link-) tag. And special params are `out` as `true` - to generate link to count ctr, `with_trade` as `true` to generate link to trade with redirection to desired page and `full_url` as `true` to generate full absolute url. Examples of using `link`:
```javascript
const url = link("content", 
  "slug", "some-content-slug", 
  "id", 12345, 
  "category", "some-category", 
  "with_trade", true,
  "full_url", true,
)
```
* `parse_ua` - function to parse user_agent (of current surfer or provided as first parameter) and return object with these properties: `URL`: url of bot, `Name`: name of browser/bot, `Version`: version of browser/bot, `OS`: name of OS, `OSVersion`: version of OS, `Device`: name of device, `Mobile`: true if mobile device, `Tablet`: true if tablet device, `Desktop`: true if desktop device, `Bot`: true if bot. Example:
```javascript
const ua = parse_ua()
if (ua.Mobile) {
  console.log("Surfer is on mobile device")
}
```

## Variables, available in template files and custom functions.

* `now` - holds current time
* `cookies` - object with surfer's cookies with field name as cookie name and field value as cookie value. Useful only with `{% dynamic %}` tag.
* `headers` - object with surfer request headers. Useful only with `{% dynamic %}` tag.
* `page_template` - page template name (`"top-categories"`, `"category"`, `"model"`, `"channel"`, `"top-content"`, `"popular"`, `"new"`, `"long"`, `"search"`, `"models"`, `"content-item"`, `"fake-player"`, `"video-embed"` or custom template name).
* `lang` - holds current page language information as [Language](Types.md#language) type.
* `ip` holds IP of surfer. Useful only with `{% dynamic %}` tag.
* `uri` holds current page URI.
* `user_agent` holds current user agent. Useful only with `{% dynamic %}` tag.
* `nocache` boolean, if true - page is requested with nocache param.
* `languages` - array of available languages, presented as Language struct, described above in `lang` variable.
* `page` - current page number.
* `host` - hostname of your site.
* `params` - object with route params.
* `query` - object with querystring params.
* `querystring` - raw querystring.
* `canonical_query` - current page canonical querystring parameters
* `config` - site configuration options. Field names are the same as in `config.toml`, but CamelCased except custom route names and custom variable names.
* `global_config` - global configuration options (in root `global-config.toml`). Field names are the same as in `config.toml`, but CamelCased.
* `route` - current route value

## Special variables, available in different template files.

In some template files there are additional variables available.
* `500.twig`:
  * `error` string - contains server error message.
* `category.twig`:
  * `category` - [requested category info](Types.md#categoryresult).
  * `content` - [ContentResults](Types.md#contentresults) for content in this category.
  * `total`, `from`, `to`, `page`, `pages` - fields from [ContentResults](Types.md#contentresults) type.
* `channel.twig`:
  * `channel` - [requested channel info](Types.md#channelresult).
  * `content` - [ContentResults](Types.md#contentresults) for content in this channel.
  * `total`, `from`, `to`, `page`, `pages` - fields from [ContentResults](Types.md#contentresults) type.
* `model.twig`:
  * `model` - [requested model info](Types.md#modelresult).
  * `content` - [ContentResults](Types.md#contentresults) for content with this model.
  * `total`, `from`, `to`, `page`, `pages` - fields from [ContentResults](Types.md#contentresults) type.
* `content-item.twig`:
  * `content_item` - [requested content item](Types.md#contentitemresult).
  * `related` - array of [ContentResult](Types.md#contentresult) with related content.
* `fake-player.twig`:
  * `content_item` - [requested video-link content info](Types.md#contentitemresult).
  * `related` - array of [ContentResult](Types.md#contentresult) with related content.
* `long.twig`:
  * `content` - [ContentResults](Types.md#contentresults) with content sorted by duration desc.
  * `total`, `from`, `to`, `page`, `pages` - fields from [ContentResults](Types.md#contentresults) type.
* `models.twig`:
  * `content` - [ModelResults](Types.md#modelresults) with requested models.
  * `total`, `from`, `to`, `page`, `pages` - fields from [ModelResults](Types.md#modelresults) type.
* `new.twig`:
  * `content` - [ContentResults](Types.md#contentresults) with content sorted by dated desc.
  * `total`, `from`, `to`, `page`, `pages` - fields from [ContentResults](Types.md#contentresults) type.
* `popular.twig`:
  * `content` - [ContentResults](Types.md#contentresults) with content sorted by monthly views desc.
  * `total`, `from`, `to`, `page`, `pages` - fields from [ContentResults](Types.md#contentresults) type.
* `search.twig`:
  * `search_query` - string with requested search query
  * `content` - [ContentResults](Types.md#contentresults) with content containing search query.
  * `total`, `from`, `to`, `page`, `pages` - fields from [ContentResults](Types.md#contentresults) type.
* `top-categories.twig`:
  * `top_categories` - [CategoryResults](Types.md#categoryresults) with top categories by CTR.
  * `total`, `from`, `to`, `page`, `pages` - fields from [CategoryResults](Types.md#categoryresults) type.
* `top-content.twig`:
  * `content` - [ContentResults](Types.md#contentresults) with content sorted by CTR.
  * `total`, `from`, `to`, `page`, `pages` - fields from [ContentResults](Types.md#contentresults) type.
* `video-embed.twig`:
  * `content_item` - [requested content item](Types.md#contentitemresult).
  * `related` - array of [ContentResult](Types.md#contentresult) with related content.

## Custom functions
You can write custom functions in javascript, which will be available in template files. Custom functions need to be located in 
`extensions/function-{function_name}.js` files. Example site has already some custom functions to inspect. In custom function you can use any functions and variables above, which available in template files. And also some other:
* `faker` - this variable is the instance of [gofaker](https://github.com/brianvoe/gofakeit) and contains 
these [methods](https://pkg.go.dev/github.com/brianvoe/gofakeit). It is useful if you need to make some stub for development environment. 
* `cache(cacheKey string, timeout string, recreate func())` - this function can be used to fetch some string content with caching. Function accepts 3 parameters. First is the cache key (string), 
second - timeout duration in human readable format like `"1 hour"`, third is the function which will do cache recreating, this function must return string. 
* `fetch(url string)` - this function can be used to fetch data from any URL. It has chaining style. `fetch()` initializes with URL of your choice as first argument and 
returns `fetchRequest` object with these methods:
  * `WithMethod(method string)` - adds request method (`GET`, `POST`, `PUT` etc) and returns updated `fetchRequest` object.
  * `WithUrl(url string)` - rewrite url on initialization and returns updated `fetchRequest` object.
  * `WithHeader(headerName string, headerValue string)` - adds header to request and returns updated `fetchRequest` object.
  * `WithHeaders(headers object)` - adds headers as object with names as field names and values as field values and returns updated `fetchRequest` object.
  * `WithQueryParam(paramName string, paramValue string)` - adds querystring param to URL and returns updated `fetchRequest` object.
  * `WithQueryString(querystring string)` - sets raw querystring for URL and returns updated `fetchRequest` object.
  * `WithRawData(data []byte)` - sets raw body data for POST or PUT requests in bytes and returns updated `fetchRequest` object.
  * `WithJsonData(data any)` - sets JSON body data for POST or PUT requests as any object and returns updated `fetchRequest` object. It also sets header `Content-Type` to `application/json`
  * `WithFormData(data object)` - sets Form body data with data in passed object for POST or PUT requests and returns updated `fetchRequest` object. It also sets header `Content-Type` to `application/x-www-form-urlencoded;charset=UTF-8` 
  * `Json()` do JSON API request and return result as object with JSON data or null in case of error.
  * `Raw()` do request and return result as raw data in bytes or null in case of error.
  * `String()` do request and return result as string or empty string in case of error.
  Example:
  ```javascript
    // this function will return IP of your server after 1 second and cache it on subsequent requests.
    function example_function() {
      return cache("test", "1 minute", function () {
        const res = fetch("http://httpbin.org/delay/1").WithMethod("POST").WithJsonData({some_var: "some_value"}).Json()
        if (res) {
          // httpbin returns origin IP address in "origin" field
          return res.origin
        } else {
          console.log("Some error occured fetching from httpbin")
          return "" // some error occured
        }
      })
    }
  ```
  Using `fetch()` function you can easily use other scripts in different languages, like `PHP` to get some other data. Just pass to your other script some data using `POST` or `GET` request and get output data.

## Site javascript build system

In `js` path located javascript build files tree, based on npm/yarn packaging. in the root you can create entries like `main.ts` (in [Typescript](https://www.typescriptlang.org/) language), `video.js`, etc. 
Each entry will become the ready built file with same name and with `.js`extension in `public` folder (or in folder under `public` as configured in [Site Configuration](#site-configuration)). If you imported `css` files from your entries, it also will be in same path with same name and extension `.css`. In entry files you can use any [NPM js](https://www.npmjs.com/) packages. To start working with your `js` build system you need to initialize it first:
1. Download and install [nodejs](https://nodejs.org/en/download/).
2. In command change directory to your `js` path and run:
```shell
npm install
```
Now you can install any new package from [NPM js](https://www.npmjs.com/) with command `npm install <package name>` from your `js` path. And now, if you change any entry file in `js` folder, script automatically rebuilds javascript and copy new result file to `public` folder. If any error occurs, script will output them in standard output in dev mode and to standard logging system in production mode. So, when you start ./totaltube-frontend in dev mode on `windows`, just look in the window you started it to see any errors. In `linux` you can see errors with command `journalctl -u totaltube-frontend -f`. The building is very fast and takes tens of milliseconds. 

## Site css build system

For building css for site, we use [Sass](https://sass-lang.com/documentation/). Sass is a stylesheet language that’s compiled to CSS. It allows you to use variables, nested rules, mixins, functions, and more, all with a fully CSS-compatible syntax. Sass helps keep large stylesheets well-organized and makes it easy to share design within and across projects. 

All sass files are located in `scss` folder. Entry `scss` files are configured in [Site Configuration](#site-configuration). The result built `css` file with same name as entry file will be copied to `public` folder or to the [configured](#site-configuration) destination. Css will be built on each change of any `.scss` files. Errors, like with js building, will be in standard output for dev mode on `windows` and in standard logging system in production mode on `linux`. The building is very fast and takes tens of milliseconds.