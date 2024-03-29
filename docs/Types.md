# Special Totaltube frontend types

### ContentResults
this type is returned on fetching of content from Totaltube minion API. Type fields:
* `Total` integer - total amount of content
* `From` integer - from index
* `To` integer - to index
* `Page` integer - current page number
* `Pages` integer - total number of pages
* `Items` array of [ContentResult](#contentresult) - content items
* `Title` string - if the content is filtered by category, model or channel, this title will contain title of corresponding category, model or channel.
* `Related` array of [RelatedItem](#relateditem) - related categories, models, channels or search queries if the content is filtered by category, model, channel or search query.

### ContentResult
Type fields:
* `Id` integer - numeric ID of content
* `Title` string - content title
* `TitleTranslated` boolean - if true, the `Title` is translated to requested language.
* `OriginalTitle` string - if title is translated, this field holds original title
* `Description` string - content description
* `Channel` [ChannelShortResult](#channelshortresult) - content channel
* `Link` string - for `video-link` type this field will contain link to video
* `CreatedAt` time.Time - actual content creation time
* `Dated` time.Time - appointed content creation time
* `Duration` [ContentDuration](#contentduration) - content duration in seconds
* `Tags` array of strings - content tags
* `Keywords` array of strings - content keywords
* `Type` string - type of content. One of (`"video-embed"`, `"video-link"`, `"video"`, `"gallery"`)
* `Priority` integer - priority for content creation and rotation.
* `User` [ContentResultUser](#contentresultuser) - content creator
* `Categories` array of [TaxonomyResult](#taxonomyresult) - content categories
* `Models` array of [TaxonomyResult](#taxonomyresult) - content models
* `RotationStatus` string - for 1 page of category top or top content page it will be `"casting"` o `"normal"` indicating if the content is in casting state or already got CTR counted.
* `Ctr` float - current content CTR (for category top or top content page).
* `Views` integer - content views (by default for last month).
* `SourceSiteId` string - id of source site if grabbed using content sources.
* `SourceSiteUniqueId` string - unique ID of this content on source site if grabbed using content sources.
* `GetThumbFormat(formatName? string)` [ThumbFormat](#thumbformat) - function returns [ThumbFormat](#thumbformat) with name in optional argument or first available format.
* `ThumbTemplate(formatName? string)` string - function return thumb template URL with %d on the place of thumb number.
* `Thumb(formatName? string)` string - function return thumb URL.
* `HiresThumb(formatName? string)` string - function return hires thumb or ordinary thumb if no hires thumb available.
* `SelectedThumb(formatName? string)` integer - returns the index of currently selected thumb to show if content has several thumbs.
* `MainCategorySlug(defaultName string)` string - returns the slug of content category or `defaultName` if content not in any category. Useful to generate links to content with category slug in link.
* `HasCustomField(fieldName string)` bool - returns true if content item has custom field with name `fieldName`
* `CustomField(fieldName string)` any - returns the value of custom field with name `fieldName` if it's defined, or `null`
* `HasCustomTranslation(key string)` bool - returns true if content item has custom translation in current language for key `key`
* `CustomTranslation(key string)` string - returns the custom translation with key `key`

### ThumbFormat
This type holds some information about thumb format
* `Name` string - the name of format
* `Amount` integer - amount of thumbs of this format
* `Width` integer - width of thumb
* `Height` integer - height of thumb
* `Type` string - thumb type (`png`, `jpg`, `webp`)
* `Retina` boolean - if true, thumb has highres "retina" version.

### ContentItemResult
This type has the same fields as in [ContentResult](#contentresult), with this additional fields:
* `Related` array of [ContentResult](#contentresult) - array of similar content
* `GalleryFormats()` array of strings - function returns all gallery format names for content of type `"gallery"`
* `GalleryImages(galleryFormat string)` array of [GalleryImageInfo](#galleryimageinfo) - function returns gallery images information for rendering image gallery (only for content type `"gallery"`). First argument is optional and can be the name of gallery format.
* `VideoInfo(videoFormat string)` [ContentVideoInfo](#contentvideoinfo) - function returns information about video format for content of type `"video"`. First argument is optional and can be the name of video format.
* `VideoFormats()` array of strings - function returns all video format names for content of type `"video"`
* `VideoUrl(videoFormat? string)` string - function returns video url for content of type `"video"`. `videoFormat` argument is optional.
* `VideoPoster(videoFormat? string)` string - function returns video poster image URL for content of type `"video"`. `videoFormat` argument is optional. If no `videoFormat` specified, the function will return poster for format with maximum video size.
* `VideoTimeline(videoFormat? string)` string - function returns video timeline `.vtt` file URL for content of type `"video"`. `videoFormat` argument is optional.
* `VideoSize(videoFormat? string)` [Size](#size) - function returns video size for content of type `"video"`. `videoFormat` argument is optional.
* `MaxVideoSize()` [Size](#size) - function returns maximum video size among all formats for content of type `"video"`.
* `Mp4VideoFormats()` array of strings - function returns all video format names for content of type `"video"` with type `"mp4"`.
* `HlsMasterUrl()` string - function returns HLS master playlist URL for content of type `"video"` if you have [nginx-vod](https://github.com/kaltura/nginx-vod-module) setup.

### RelatedItem
The type has the following fields:
* `Message` string - title of related taxonomy or search query
* `Type` string - type of related item (`"category"`, `"model"`, `"channel"` or empty for search query)
* `Id` integer - numeric ID of related taxonomy if the `Type` is `"category"`, `"model"` or `"channel"`. 0 for search query.
* `Slug` string - slug of related taxonomy if the `Type` is `"category"`, `"model"` or `"channel"`. Empty for search query.

### TaxonomyResult
The type has the following fields:
* `Id` integer - numeric ID of category, model or channel
* `Slug` string - slug of category, model or channel
* `Title` string - title of category, model or channel

### ChannelShortResult
The type has the following fields:
* `Id` integer - numeric ID of channel
* `Slug` string - slug of channel
* `Title` string - title of channel
* `Url` string - url of channel
* `Banner` string - banner of channel

### ContentDuration
This type itself is integer, holds duration in seconds and can be used directly as integer. Also, this type has these functions:
* `Format()` string - function returns duration in `mm:ss` format.

### ContentResultUser
The type has the following fields:
* `Id` integer - numeric ID of user.
* `Login` string - login of user.
* `Name` string - name of user.

### GalleryImageInfo
The type has the following fields:
* `ImageUrl` string - big image URL
* `PreviewUrl` string - preview image URL
* `PreviewSize` [Size](#size) - preview image size
* `ImageSize` [Size](#size) - big image size

### Size
The type has the following Fields:
* `Width` integer
* `Height` integer

### ContentVideoInfo
The type has the following fields:
* `Name` string - video format name
* `Type` string - video type (`"mp4"` or `"webm"`)
* `Size` [Size](#size) - video frame size
* `VideoBitrate` integer - video bitrate in bytes
* `AudioBitrate` integer - audio bitrate in bytes
* `PosterType` string - poster image type (`"jpg"`, `"webp"` or `"png"`)
* `TimelineType` string - timeline image type (`"jpg"`, `"webp"` or `"png"`)
* `TimelineSize` [Size](#size) - the size of timeline image
* `TimelineFrames` integer - the amount of timeline frames
* `Duration` float - video duration

### CategoryResults
This type is returned on fetching categories. The type has the following fields:
* `Total` integer - total number of categories matching the search criteria.
* `From` integer - from index (for pagination)
* `To` integer - to index (for pagination)
* `Page` integer - current page
* `Pages` integer - total number of pages
* `Items` array of [CategoryResult](#categoryresult) - category items

### CategoryResult
The type has the following fields:
* `Id` integer - numeric ID of category
* `Slug` string - slug of category
* `Title` string - title of category
* `TitleTranslated` boolean - true if category title has translation.
* `OriginalTitle` string - if title translated, it holds original title
* `Description` string - category description
* `Tags` array of strings - category tags
* `Dated` time.Time - category dated time
* `CreatedAt` time.Time - category actual creation time
* `AliasCategoryId` integer - if category has alias - here is the alias category ID
* `RotationStatus` string - category rotation status for top categories page. Can be `"casting"` or `"normal"`.
* `Total` integer - total amount of content in category.
* `Ctr` float - category CTR.
* `Views` integer - category views for last month.
* `GetThumbFormat(formatName? string)` [ThumbFormat](#thumbformat) - function returns [ThumbFormat](#thumbformat) with name in optional argument or first available format.
* `ThumbTemplate(formatName? string)` string - function return thumb template URL with %d on the place of thumb number.
* `Thumb(formatName? string)` string - function return thumb URL.
* `HiresThumb(formatName? string)` string - function return hires thumb or ordinary thumb if no hires thumb available.
* `SelectedThumb(formatName? string)` integer - returns the index of currently selected thumb to show if category has several thumbs.
* `HasCustomField(fieldName string)` bool - returns true if category has custom field with name `fieldName`
* `CustomField(fieldName string)` any - returns the value of custom field with name `fieldName` if it's defined, or `null`
* `HasCustomTranslation(key string)` bool - returns true if category has custom translation in current language for key `key`
* `CustomTranslation(key string)` string - returns the custom translation with key `key`

### ModelResults
This type is returned on fetching models. The type has the following fields:
* `Total` integer - total number of models matching the search criteria.
* `From` integer - from index (for pagination)
* `To` integer - to index (for pagination)
* `Page` integer - current page
* `Pages` integer - total number of pages
* `Items` array of [ModelResult](#modelresult) - model items

### ModelResult
The type has the following fields:
* `Id` integer - numeric ID of model
* `Slug` string - slug of model
* `Title` string - title of model
* `TitleTranslated` boolean - true if model title has translation.
* `OriginalTitle` string - if title translated, it holds original title
* `Description` string - model description
* `Tags` array of strings - model tags
* `Dated` time.Time - model dated time
* `CreatedAt` time.Time - model actual creation time
* `Total` integer - total amount of content with this model.
* `Views` integer - model views for last month.
* `GetThumbFormat(formatName? string)` [ThumbFormat](#thumbformat) - function returns [ThumbFormat](#thumbformat) with name in optional argument or first available format.
* `ThumbTemplate(formatName? string)` string - function return thumb template URL with %d on the place of thumb number.
* `Thumb(formatName? string)` string - function return thumb URL.
* `HiresThumb(formatName? string)` string - function return hires thumb or ordinary thumb if no hires thumb available.
* `SelectedThumb(formatName? string)` integer - returns the index of currently selected thumb to show if model has several thumbs.
* `HasCustomField(fieldName string)` bool - returns true if model has custom field with name `fieldName`
* `CustomField(fieldName string)` any - returns the value of custom field with name `fieldName` if it's defined, or `null`
* `HasCustomTranslation(key string)` bool - returns true if model has custom translation in current language for key `key`
* `CustomTranslation(key string)` string - returns the custom translation with key `key`


### ChannelResults
This type is returned on fetching channels. The type has the following fields:
* `Total` integer - total number of channels matching the search criteria.
* `From` integer - from index (for pagination)
* `To` integer - to index (for pagination)
* `Page` integer - current page
* `Pages` integer - total number of pages
* `Items` array of [ChannelResult](#channelresult) - channel items

### ChannelResult
The type has the following fields:
* `Id` integer - numeric ID of channel
* `Slug` string - slug of channel
* `Title` string - title of channel
* `TitleTranslated` boolean - true if channel title has translation.
* `OriginalTitle` string - if title translated, it holds original title
* `Description` string - channel description
* `Tags` array of strings - channel tags
* `Url` string - url of channel if set
* `Banner` string - banner of channel if set
* `Dated` time.Time - channel dated time
* `CreatedAt` time.Time - channel actual creation time
* `Total` integer - total amount of content in this channel.
* `Views` integer - channel views for last month.
* `GetThumbFormat(formatName? string)` [ThumbFormat](#thumbformat) - function returns [ThumbFormat](#thumbformat) with name in optional argument or first available format.
* `ThumbTemplate(formatName? string)` string - function return thumb template URL with %d on the place of thumb number.
* `Thumb(formatName? string)` string - function return thumb URL.
* `HiresThumb(formatName? string)` string - function return hires thumb or ordinary thumb if no hires thumb available.
* `SelectedThumb(formatName? string)` integer - returns the index of currently selected thumb to show if model has several thumbs.
* `HasCustomField(fieldName string)` bool - returns true if channel has custom field with name `fieldName`
* `CustomField(fieldName string)` any - returns the value of custom field with name `fieldName` if it's defined, or `null`
* `HasCustomTranslation(key string)` bool - returns true if channel has custom translation in current language for key `key`
* `CustomTranslation(key string)` string - returns the custom translation with key `key`

### TopSearch
The type has the following fields:
* `Message` string - search query 
* `Searches` integer - number of searches

### Language
The type has the following fields:
* `Id` integer - language ID like "en", "de", "it"
* `Name` string - language name like English, German, Italian
* `Locale` string - language locale like "en_US", "de_DE", "it_IT"
* `Native` string - native language name like English, Deutsch, Italiano
* `Direction` string - language direction: "ltr" (left to right) or "rtl" (right to left)
* `Country` string - country code associated with language like us, de, it