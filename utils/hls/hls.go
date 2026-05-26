package hls

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type Quality struct {
	Name      string
	Width     int
	Height    int
	Bitrate   string
	AudioRate string
}

var DefaultQualities = []Quality{
	{Name: "1080p", Width: 1920, Height: 1080, Bitrate: "5000k", AudioRate: "192k"},
	{Name: "720p", Width: 1280, Height: 720, Bitrate: "2500k", AudioRate: "128k"},
	{Name: "480p", Width: 854, Height: 480, Bitrate: "1000k", AudioRate: "128k"},
	{Name: "360p", Width: 640, Height: 360, Bitrate: "600k", AudioRate: "96k"},
}

type HLSResult struct {
	MasterPlaylistPath string
	QualityDirs        map[string]string
}

func GenerateHLS(inputPath, outputDir string, qualities []Quality, baseURL string) (*HLSResult, error) {
	os.MkdirAll(outputDir, 0755)
	result := &HLSResult{
		QualityDirs: make(map[string]string),
	}

	for _, q := range qualities {
		qualityDir := filepath.Join(outputDir, q.Name)
		os.MkdirAll(qualityDir, 0755)

		playlistPath := filepath.Join(qualityDir, "playlist.m3u8")
		segmentPattern := filepath.Join(qualityDir, "seg%03d.ts")

		cmd := exec.Command(
			"ffmpeg",
			"-i", inputPath,
			"-vf", fmt.Sprintf("scale=%d:%d", q.Width, q.Height),
			"-vcodec", "libx264",
			"-b:v", q.Bitrate,
			"-maxrate", q.Bitrate,
			"-bufsize", doubleRate(q.Bitrate),
			"-preset", "faster",
			"-g", "48",
			"-sc_threshold", "0",
			"-acodec", "aac",
			"-b:a", q.AudioRate,
			"-ar", "44100",
			"-f", "hls",
			"-hls_time", "6",
			"-hls_list_size", "0",
			"-hls_segment_filename", segmentPattern,
			playlistPath,
		)

		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		cmd.Stdout = os.Stdout
		if err := cmd.Run(); err != nil {
			return nil, fmt.Errorf("ffmpeg %s failed: %s", q.Name, stderr.String())
		}

		result.QualityDirs[q.Name] = qualityDir
	}

	masterPath := filepath.Join(outputDir, "master.m3u8")
	if err := generateMasterPlaylist(masterPath, qualities, baseURL); err != nil {
		return nil, err
	}
	result.MasterPlaylistPath = masterPath

	return result, nil
}

func generateMasterPlaylist(outputPath string, qualities []Quality, baseURL string) error {
	content := "#EXTM3U\n#EXT-X-VERSION:3\n\n"
	for _, q := range qualities {
		bandwidth := bitrateToInt(q.Bitrate) + bitrateToInt(q.AudioRate)
		content += fmt.Sprintf(
			"#EXT-X-STREAM-INF:BANDWIDTH=%d,RESOLUTION=%dx%d,NAME=\"%s\"\n%s/%s/playlist.m3u8\n\n",
			bandwidth, q.Width, q.Height, q.Name, baseURL, q.Name,
		)
	}
	return os.WriteFile(outputPath, []byte(content), 0644)
}

func doubleRate(rate string) string {
	var n int
	fmt.Sscanf(rate, "%dk", &n)
	return fmt.Sprintf("%dk", n*2)
}

func bitrateToInt(rate string) int {
	var n int
	fmt.Sscanf(rate, "%dk", &n)
	return n * 1000
}
