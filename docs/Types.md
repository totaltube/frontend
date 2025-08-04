# Special Totaltube frontend types

This document describes the data types available in the Totaltube frontend system. These types are used throughout the application for data manipulation and template rendering.

### ContentResults
This type is returned when fetching content from the Totaltube minion API. It contains the following fields:
* `Total` integer - total amount of content items matching the query criteria
* `From` integer - starting index of the current page
* `To` integer - ending index of the current page
* `Page` integer - current page number
* `Pages` integer - total number of available pages
* `Items` array of [ContentResult](#contentresult) - collection of content items
* `Title` string - title of the category, model, or channel if content is filtered by these criteria
* `Related` array of [RelatedItem](#relateditem) - related categories, models, channels, or search queries based on current filtering


### ContentResult
This type represents an individual content item with the following fields:
* `Id` integer - unique numeric identifier of the content
* `Title` string - content title
* `TitleTranslated` boolean - indicates if the title has been translated to the requested language
* `OriginalTitle` string - original title if translation is applied
* `Description` string - content description
* `Channel` [ChannelShortResult](#channelshortresult) - channel associated with the content
* `Link` string - external link for content of type `"video-link"`
* `CreatedAt` time.Time - actual content creation timestamp
* `Dated` time.Time - appointed content publication timestamp
* `Duration` [ContentDuration](#contentduration) - content duration in seconds
* `Tags` array of strings - content tags
* `Keywords` array of strings - content keywords
* `Type` string - content type, one of: `"video-embed"`, `"video-link"`, `"video"`, `"gallery"`
* `Priority` integer - content priority for creation and rotation processes
* `User` [ContentResultUser](#contentresultuser) - content creator information
* `Categories` array of [TaxonomyResult](#taxonomyresult) - categories associated with the content
* `Models` array of [TaxonomyResult](#taxonomyresult) - models featured in the content
* `RotationStatus` string - content rotation status on category top or top content pages; can be `"casting"` or `"normal"`
* `Ctr` float - current click-through rate (for category top or top content pages)
* `Views` integer - view count (default: last month's views)
* `SourceSiteId` string - source site identifier if content was obtained via content sources
* `SourceSiteUniqueId` string - unique identifier of this content on the source site
* `CustomData` [CustomData](#customdata) - custom fields data associated with this content
* `CustomTranslations` [CustomTranslations](#customtranslations) - custom translations for this content

**Methods:**
* `GetThumbFormat(formatName? string)` [ThumbFormat](#thumbformat) - returns thumbnail format by name or the first available format
* `ThumbTemplate(formatName? string)` string - returns thumbnail template URL with `%d` placeholder for thumbnail number
* `Thumb(formatName? string)` string - returns complete thumbnail URL
* `HiresThumb(formatName? string)` string - returns high-resolution thumbnail URL or standard thumbnail if unavailable
* `SelectedThumb(formatName? string)` integer - returns the index of the currently selected thumbnail
* `MainCategorySlug(defaultName? string)` string - returns the slug of the content's main category or the default value if no category exists
* `HasCustomField(fieldName string)` boolean - checks if a specific custom field exists
* `CustomField(fieldName string)` any - returns the value of the specified custom field or null
* `HasCustomTranslation(key string)` boolean - checks if a custom translation exists for the given key
* `CustomTranslation(key string)` string - returns the custom translation for the given key


### ThumbFormat
This type contains information about a thumbnail format:
* `Name` string - format name
* `Amount` integer - number of thumbnails available in this format
* `Width` integer - thumbnail width in pixels
* `Height` integer - thumbnail height in pixels
* `Type` string - thumbnail image format (`"png"`, `"jpg"`, `"webp"`)
* `Retina` boolean - indicates if high-resolution "retina" versions are available

### ContentItemResult
This type extends [ContentResult](#contentresult) with additional fields for individual content items:
* `Related` array of [ContentResult](#contentresult) - collection of related content items

**Additional Methods:**
* `GalleryFormats()` array of strings - returns all available gallery format names for gallery-type content
* `GalleryImages(galleryFormat? string)` array of [GalleryImageInfo](#galleryimageinfo) - returns gallery image information for rendering
* `VideoInfo(videoFormat? string)` [ContentVideoInfo](#contentvideoinfo) - returns video format information
* `VideoFormats()` array of strings - returns all available video format names
* `VideoUrl(videoFormat? string)` string - returns video file URL
* `VideoPoster(videoFormat? string)` string - returns video poster image URL
* `VideoTimeline(videoFormat? string)` string - returns video timeline `.vtt` file URL
* `VideoSize(videoFormat? string)` [Size](#size) - returns video dimensions
* `MaxVideoSize()` [Size](#size) - returns the largest available video dimensions
* `Mp4VideoFormats()` array of strings - returns names of available MP4 video formats
* `HlsMasterUrl()` string - returns HLS master playlist URL (requires [nginx-vod](https://github.com/kaltura/nginx-vod-module))

### RelatedItem
This type represents related content references:
* `Message` string - title of related taxonomy or search query
* `Type` string - type of relationship: `"category"`, `"model"`, `"channel"` or empty for search query
* `Id` integer - numeric identifier of the related taxonomy (0 for search query)
* `Slug` string - slug of the related taxonomy (empty for search query)
* `Searches` integer - number of searches (only for search queries)

### TaxonomyResult
This type represents a category, model, or channel reference:
* `Id` integer - numeric identifier
* `Slug` string - URL-friendly slug
* `Title` string - display title

### ChannelShortResult
This type provides basic channel information:
* `Id` integer - channel identifier
* `Slug` string - channel slug
* `Title` string - channel title
* `Url` string - channel URL
* `Banner` string - channel banner image

### ContentDuration
This integer type represents duration in seconds and provides formatting methods:
* `Format()` string - returns duration in `mm:ss` format

### ContentResultUser
This type represents content creator information:
* `Id` integer - user identifier
* `Login` string - user login name
* `Name` string - user display name

### GalleryImageInfo
This type provides information about gallery images:
* `ImageUrl` string - full-size image URL
* `PreviewUrl` string - thumbnail image URL
* `PreviewSize` [Size](#size) - thumbnail dimensions
* `ImageSize` [Size](#size) - full-size image dimensions

### Size
This type represents image or video dimensions:
* `Width` integer - width in pixels
* `Height` integer - height in pixels

### ContentVideoInfo
This type provides video format information:
* `Name` string - video format name
* `Type` string - video format type (`"mp4"` or `"webm"`)
* `Size` [Size](#size) - video frame dimensions
* `VideoBitrate` integer - video bitrate in bytes per second
* `AudioBitrate` integer - audio bitrate in bytes per second
* `PosterType` string - poster image format (`"jpg"`, `"webp"`, or `"png"`)
* `TimelineType` string - timeline image format (`"jpg"`, `"webp"`, or `"png"`)
* `TimelineSize` [Size](#size) - timeline image dimensions
* `TimelineFrames` integer - number of timeline frames
* `Duration` float - video duration in seconds

### CategoryResults
This type is returned when fetching categories:
* `Total` integer - total number of matching categories
* `From` integer - starting index for pagination
* `To` integer - ending index for pagination
* `Page` integer - current page number
* `Pages` integer - total number of pages
* `Items` array of [CategoryResult](#categoryresult) - collection of category items

### CategoryResult
This type represents a content category:
* `Id` integer - category identifier
* `Slug` string - category slug
* `Title` string - category title
* `TitleTranslated` boolean - indicates if the title is translated
* `OriginalTitle` string - original title if translation is applied
* `Description` string - category description
* `Tags` array of strings - category tags
* `Dated` time.Time - category publication date
* `CreatedAt` time.Time - category creation date
* `AliasCategoryId` integer - identifier of aliased category, if applicable
* `RotationStatus` string - rotation status for top categories page (`"casting"` or `"normal"`)
* `Total` integer - total number of content items in this category
* `Ctr` float - category click-through rate
* `Views` integer - view count (default: last month)
**Methods:**
* `GetThumbFormat(formatName? string)` [ThumbFormat](#thumbformat) - returns specified thumbnail format or first available
* `ThumbTemplate(formatName? string)` string - returns thumbnail template URL
* `Thumb(formatName? string)` string - returns thumbnail URL
* `HiresThumb(formatName? string)` string - returns high-resolution thumbnail URL
* `SelectedThumb(formatName? string)` integer - returns index of selected thumbnail
* `HasCustomField(fieldName string)` boolean - checks for custom field existence
* `CustomField(fieldName string)` any - retrieves custom field value
* `HasCustomTranslation(key string)` boolean - checks for custom translation existence
* `CustomTranslation(key string)` string - retrieves custom translation

### ModelResults
This type is returned when fetching models:
* `Total` integer - total number of matching models
* `From` integer - starting index for pagination
* `To` integer - ending index for pagination
* `Page` integer - current page number
* `Pages` integer - total number of pages
* `Items` array of [ModelResult](#modelresult) - collection of model items

### ModelResult
This type represents a content model:
* `Id` integer - model identifier
* `Slug` string - model slug
* `Title` string - model title
* `TitleTranslated` boolean - indicates if the title is translated
* `OriginalTitle` string - original title if translation is applied
* `Description` string - model description
* `Tags` array of strings - model tags
* `Dated` time.Time - model publication date
* `CreatedAt` time.Time - model creation date
* `Total` integer - total number of content items featuring this model
* `Views` integer - view count (default: last month)

**Methods:**
* `GetThumbFormat(formatName? string)` [ThumbFormat](#thumbformat) - returns specified thumbnail format or first available
* `ThumbTemplate(formatName? string)` string - returns thumbnail template URL
* `Thumb(formatName? string)` string - returns thumbnail URL
* `HiresThumb(formatName? string)` string - returns high-resolution thumbnail URL
* `SelectedThumb(formatName? string)` integer - returns index of selected thumbnail
* `HasCustomField(fieldName string)` boolean - checks for custom field existence
* `CustomField(fieldName string)` any - retrieves custom field value
* `HasCustomTranslation(key string)` boolean - checks for custom translation existence
* `CustomTranslation(key string)` string - retrieves custom translation

### ChannelResults
This type is returned when fetching channels:
* `Total` integer - total number of matching channels
* `From` integer - starting index for pagination
* `To` integer - ending index for pagination
* `Page` integer - current page number
* `Pages` integer - total number of pages
* `Items` array of [ChannelResult](#channelresult) - collection of channel items

### ChannelResult
This type represents a content channel:
* `Id` integer - channel identifier
* `Slug` string - channel slug
* `Title` string - channel title
* `TitleTranslated` boolean - indicates if the title is translated
* `OriginalTitle` string - original title if translation is applied
* `Description` string - channel description
* `Tags` array of strings - channel tags
* `Url` string - channel URL, if configured
* `Banner` string - channel banner image, if configured
* `Dated` time.Time - channel publication date
* `CreatedAt` time.Time - channel creation date
* `Total` integer - total number of content items in this channel
* `Views` integer - view count (default: last month)

**Methods:**
* `GetThumbFormat(formatName? string)` [ThumbFormat](#thumbformat) - returns specified thumbnail format or first available
* `ThumbTemplate(formatName? string)` string - returns thumbnail template URL
* `Thumb(formatName? string)` string - returns thumbnail URL
* `HiresThumb(formatName? string)` string - returns high-resolution thumbnail URL
* `SelectedThumb(formatName? string)` integer - returns index of selected thumbnail
* `HasCustomField(fieldName string)` boolean - checks for custom field existence
* `CustomField(fieldName string)` any - retrieves custom field value
* `HasCustomTranslation(key string)` boolean - checks for custom translation existence
* `CustomTranslation(key string)` string - retrieves custom translation

### TopSearch
This type represents a popular search query:
* `Message` string - search query text
* `Searches` integer - number of times the query was searched

### Language
This type represents a supported language:
* `Id` string - language code (e.g., "en", "de", "it")
* `Name` string - language name in English (e.g., "English", "German", "Italian")
* `Locale` string - language locale (e.g., "en_US", "de_DE", "it_IT")
* `Native` string - language name in its native script (e.g., "English", "Deutsch", "Italiano")
* `Direction` string - text direction: "ltr" (left to right) or "rtl" (right to left)
* `Country` string - associated country code (e.g., "us", "de", "it")

### CountryGroup
This type represents a country group:
* `Id` integer - country group identifier
* `Name` string - country group name
* `Countries` array of strings - countries in this group
* `Ignore` boolean - if true, the country group is ignored


### CustomData
This type represents a map of custom fields associated with content items, categories, models, or channels:
* Key-value pairs where keys are field names and values can be of any type

### CustomTranslations
This type represents a map of custom translations for content items, categories, models, or channels:
* Key-value pairs where keys are translation identifiers and values are localized strings