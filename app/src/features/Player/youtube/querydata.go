package youtube

type QueryData struct {
	VideoInfo *struct {
		Filename string `json:"filename"`
		Id       string `json:"id"`
		Title    string `json:"title"`
		Formats  []struct {
			FormatId   string  `json:"format_id"`
			FormatNote string  `json:"format_note,omitempty"`
			Ext        string  `json:"ext"`
			Protocol   string  `json:"protocol"`
			Acodec     string  `json:"acodec,omitempty"`
			Vcodec     string  `json:"vcodec"`
			Url        string  `json:"url"`
			Width      int     `json:"width,omitempty"`
			Height     int     `json:"height,omitempty"`
			Fps        float64 `json:"fps,omitempty"`
			Rows       int     `json:"rows,omitempty"`
			Columns    int     `json:"columns,omitempty"`
			Fragments  []struct {
				Url      string  `json:"url"`
				Duration float64 `json:"duration"`
			} `json:"fragments,omitempty"`
			Resolution  string  `json:"resolution"`
			AspectRatio float64 `json:"aspect_ratio"`
			HttpHeaders struct {
			} `json:"http_headers"`
			AudioExt           string      `json:"audio_ext"`
			VideoExt           string      `json:"video_ext"`
			Vbr                float64     `json:"vbr"`
			Abr                float64     `json:"abr"`
			Tbr                float64     `json:"tbr"`
			Format             string      `json:"format"`
			FormatIndex        interface{} `json:"format_index"`
			ManifestUrl        string      `json:"manifest_url,omitempty"`
			Language           string      `json:"language,omitempty"`
			Preference         int         `json:"preference,omitempty"`
			Quality            float64     `json:"quality,omitempty"`
			HasDrm             bool        `json:"has_drm,omitempty"`
			SourcePreference   int         `json:"source_preference,omitempty"`
			Asr                int         `json:"asr,omitempty"`
			Filesize           int         `json:"filesize,omitempty"`
			AudioChannels      int         `json:"audio_channels,omitempty"`
			LanguagePreference int         `json:"language_preference,omitempty"`
			DynamicRange       string      `json:"dynamic_range,omitempty"`
			Container          string      `json:"container,omitempty"`
			DownloaderOptions  struct {
				HttpChunkSize int `json:"http_chunk_size"`
			} `json:"downloader_options,omitempty"`
			FilesizeApprox int `json:"filesize_approx,omitempty"`
		} `json:"formats"`
		Thumbnails []struct {
			Url        string `json:"url"`
			Preference int    `json:"preference"`
			Id         string `json:"id"`
			Height     int    `json:"height,omitempty"`
			Width      int    `json:"width,omitempty"`
			Resolution string `json:"resolution,omitempty"`
		} `json:"thumbnails"`
		Thumbnail         string               `json:"thumbnail"`
		Description       string               `json:"description"`
		ChannelId         string               `json:"channel_id"`
		ChannelUrl        string               `json:"channel_url"`
		Duration          int                  `json:"duration"`
		ViewCount         int                  `json:"view_count"`
		AverageRating     interface{}          `json:"average_rating"`
		AgeLimit          int                  `json:"age_limit"`
		WebpageUrl        string               `json:"webpage_url"`
		Categories        []string             `json:"categories"`
		Tags              []string             `json:"tags"`
		PlayableInEmbed   bool                 `json:"playable_in_embed"`
		LiveStatus        string               `json:"live_status"`
		ReleaseTimestamp  interface{}          `json:"release_timestamp"`
		FormatSortFields  []string             `json:"_format_sort_fields"`
		AutomaticCaptions map[string][]Caption `json:"automatic_captions"`
		Subtitles         map[string][]Caption `json:"subtitles"`
		CommentCount      int                  `json:"comment_count"`
		Chapters          interface{}          `json:"chapters"`
		Heatmap           []struct {
			StartTime float64 `json:"start_time"`
			EndTime   float64 `json:"end_time"`
			Value     float64 `json:"value"`
		} `json:"heatmap"`
		LikeCount            int    `json:"like_count"`
		Channel              string `json:"channel"`
		ChannelFollowerCount int    `json:"channel_follower_count"`
		ChannelIsVerified    bool   `json:"channel_is_verified"`
		Uploader             string `json:"uploader"`
		UploaderId           string `json:"uploader_id"`
		UploaderUrl          string `json:"uploader_url"`
		UploadDate           string `json:"upload_date"`
		Availability         string `json:"availability"`
		OriginalUrl          string `json:"original_url"`
		WebpageUrlBasename   string `json:"webpage_url_basename"`
		WebpageUrlDomain     string `json:"webpage_url_domain"`
		Extractor            string `json:"extractor"`
		ExtractorKey         string `json:"extractor_key"`
		PlaylistCount        int    `json:"playlist_count"`
		Playlist             string `json:"playlist"`
		PlaylistId           string `json:"playlist_id"`
		PlaylistTitle        string `json:"playlist_title"`
		PlaylistUploader     string `json:"playlist_uploader"`
		PlaylistUploaderId   string `json:"playlist_uploader_id"`
		NEntries             int    `json:"n_entries"`
		PlaylistIndex        int    `json:"playlist_index"`
		LastPlaylistIndex    int    `json:"__last_playlist_index"`
		PlaylistAutonumber   int    `json:"playlist_autonumber"`
		DisplayId            string `json:"display_id"`
		Fulltitle            string `json:"fulltitle"`
		//DurationString       string      `json:"duration_string"`
		//IsLive               bool        `json:"is_live"`
		//WasLive              bool        `json:"was_live"`
		//RequestedSubtitles   interface{} `json:"requested_subtitles"`
		//HasDrm               interface{} `json:"_has_drm"`
		//Epoch                int         `json:"epoch"`
		//RequestedFormats     []struct {
		//	Asr                int         `json:"asr"`
		//	Filesize           int         `json:"filesize"`
		//	FormatId           string      `json:"format_id"`
		//	FormatNote         string      `json:"format_note"`
		//	SourcePreference   int         `json:"source_preference"`
		//	Fps                float64         `json:"fps"`
		//	AudioChannels      int         `json:"audio_channels"`
		//	Height             int         `json:"height"`
		//	Quality            float64     `json:"quality"`
		//	HasDrm             bool        `json:"has_drm"`
		//	Tbr                float64     `json:"tbr"`
		//	Url                string      `json:"url"`
		//	Width              int         `json:"width"`
		//	Language           string      `json:"language"`
		//	LanguagePreference int         `json:"language_preference"`
		//	Preference         interface{} `json:"preference"`
		//	Ext                string      `json:"ext"`
		//	Vcodec             string      `json:"vcodec"`
		//	Acodec             string      `json:"acodec"`
		//	DynamicRange       string      `json:"dynamic_range"`
		//	Container          string      `json:"container"`
		//	DownloaderOptions  struct {
		//		HttpChunkSize int `json:"http_chunk_size"`
		//	} `json:"downloader_options"`
		//	Protocol    string  `json:"protocol"`
		//	Resolution  string  `json:"resolution"`
		//	AspectRatio float64 `json:"aspect_ratio"`
		//	HttpHeaders struct {
		//	} `json:"http_headers"`
		//	VideoExt string  `json:"video_ext"`
		//	AudioExt string  `json:"audio_ext"`
		//	Abr      float64 `json:"abr"`
		//	Vbr      float64 `json:"vbr"`
		//	Format   string  `json:"format"`
		//} `json:"requested_formats"`
		//Format         string  `json:"format"`
		//FormatId       string  `json:"format_id"`
		//Ext            string  `json:"ext"`
		//Protocol       string  `json:"protocol"`
		//Language       string  `json:"language"`
		//FormatNote     string  `json:"format_note"`
		//FilesizeApprox int     `json:"filesize_approx"`
		//Tbr            float64 `json:"tbr"`
		//Width          int     `json:"width"`
		//Height         int     `json:"height"`
		//Resolution     string  `json:"resolution"`
		//Fps            int     `json:"fps"`
		//DynamicRange   string  `json:"dynamic_range"`
		//Vcodec         string  `json:"vcodec"`
		//Vbr            float64 `json:"vbr"`
		//StretchedRatio float64 `json:"stretched_ratio"`
		//AspectRatio    float64 `json:"aspect_ratio"`
		//Acodec         string  `json:"acodec"`
		//Abr            float64 `json:"abr"`
		//Asr            int     `json:"asr"`
		//AudioChannels  int     `json:"audio_channels"`
	} `json:"video_info"`
	Error string `json:"error"`
}

type Caption struct {
	Url      string `json:"url"`
	Ext      string `json:"ext"`
	Protocol string `json:"protocol,omitempty"`
	Name     string `json:"name,omitempty"`
}
