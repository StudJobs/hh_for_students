package models

// AchievementMeta HTTP модель
type AchievementMeta struct {
	Name      string `json:"name"`
	UserUUID  string `json:"user_uuid"`
	FileName  string `json:"file_name"`
	FileType  string `json:"file_type"`
	FileSize  int64  `json:"file_size"`
	CreatedAt string `json:"created_at"`
}

// AchievementList HTTP модель
type AchievementList struct {
	Achievements []AchievementMeta `json:"achievements"`
}

// AchievementUrl HTTP модель
type AchievementUrl struct {
	URL     string            `json:"url"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers"`
}

// AchievementData HTTP модель
type AchievementData struct {
	Meta        AchievementMeta `json:"meta"`
	UploadURL   AchievementUrl  `json:"upload_url"`
	DownloadURL AchievementUrl  `json:"download_url"`
}
