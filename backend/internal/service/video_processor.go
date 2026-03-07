package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"time"
)

type VideoInfo struct {
	Width           int
	Height          int
	DurationSeconds int
	VideoCodec      string
	VideoBitrate    int64
	AudioCodec      string
	AudioBitrate    int64
}

type ffprobeOutput struct {
	Streams []struct {
		CodecType string `json:"codec_type"`
		CodecName string `json:"codec_name"`
		Width     int    `json:"width"`
		Height    int    `json:"height"`
		BitRate   string `json:"bit_rate"`
	} `json:"streams"`
	Format struct {
		Duration string `json:"duration"`
		BitRate  string `json:"bit_rate"`
	} `json:"format"`
}

func writeTempVideo(data []byte) (string, func(), error) {
	tmp, err := os.CreateTemp("", "anzuimg-video-*.bin")
	if err != nil {
		return "", nil, err
	}
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmp.Name())
		return "", nil, err
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmp.Name())
		return "", nil, err
	}
	cleanup := func() {
		_ = os.Remove(tmp.Name())
	}
	return tmp.Name(), cleanup, nil
}

func ProbeVideoInfo(parent context.Context, data []byte) (*VideoInfo, error) {
	tmpPath, cleanup, err := writeTempVideo(data)
	if err != nil {
		return nil, fmt.Errorf("create temp video failed: %w", err)
	}
	defer cleanup()

	ctx, cancel := context.WithTimeout(parent, 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "ffprobe",
		"-v", "error",
		"-print_format", "json",
		"-show_streams",
		"-show_format",
		tmpPath,
	)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ffprobe failed: %w", err)
	}

	var parsed ffprobeOutput
	if err := json.Unmarshal(out, &parsed); err != nil {
		return nil, fmt.Errorf("parse ffprobe output failed: %w", err)
	}

	info := &VideoInfo{}
	for _, stream := range parsed.Streams {
		if stream.CodecType == "video" {
			info.Width = stream.Width
			info.Height = stream.Height
			info.VideoCodec = stream.CodecName
			if br, err := strconv.ParseInt(stream.BitRate, 10, 64); err == nil {
				info.VideoBitrate = br
			}
			continue
		}
		if stream.CodecType == "audio" {
			if info.AudioCodec == "" {
				info.AudioCodec = stream.CodecName
			}
			if br, err := strconv.ParseInt(stream.BitRate, 10, 64); err == nil {
				info.AudioBitrate = br
			}
		}
	}

	if parsed.Format.Duration != "" {
		if dur, err := strconv.ParseFloat(parsed.Format.Duration, 64); err == nil && dur > 0 {
			info.DurationSeconds = int(dur + 0.5)
		}
	}

	if info.VideoBitrate <= 0 && parsed.Format.BitRate != "" {
		if br, err := strconv.ParseInt(parsed.Format.BitRate, 10, 64); err == nil {
			info.VideoBitrate = br
		}
	}

	return info, nil
}

func GenerateVideoThumbnail(parent context.Context, data []byte, width, height int) ([]byte, error) {
	tmpIn, cleanupIn, err := writeTempVideo(data)
	if err != nil {
		return nil, fmt.Errorf("create temp input video failed: %w", err)
	}
	defer cleanupIn()

	tmpOut, err := os.CreateTemp("", "anzuimg-video-thumb-*.jpg")
	if err != nil {
		return nil, fmt.Errorf("create temp output image failed: %w", err)
	}
	outPath := tmpOut.Name()
	_ = tmpOut.Close()
	defer func() {
		_ = os.Remove(outPath)
	}()

	sizeFilter := fmt.Sprintf("scale='min(%d,iw)':'min(%d,ih)':force_original_aspect_ratio=decrease", width, height)

	ctx, cancel := context.WithTimeout(parent, 20*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-y",
		"-ss", "00:00:00",
		"-i", tmpIn,
		"-frames:v", "1",
		"-vf", sizeFilter,
		"-q:v", "4",
		outPath,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("ffmpeg thumbnail failed: %w, output: %s", err, string(out))
	}

	thumb, err := os.ReadFile(outPath)
	if err != nil {
		return nil, fmt.Errorf("read video thumbnail failed: %w", err)
	}

	return thumb, nil
}
