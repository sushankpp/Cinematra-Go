package compression

import "image"

// VideoEncoder is the top-level encoder for a full video.
// Create one per video upload, call AddFrame for each frame.
type VideoEncoder struct {
	diffEncoder *FrameDiffEncoder
	Frames      []EncodedFrame
}

// NewVideoEncoder creates a video encoder.
// quality: 1-100 (higher = better quality, less compression)
// iFrameRate: encode a full I-frame every N frames (30 = once per second at 30fps)
func NewVideoEncoder(quality, iFrameRate int) *VideoEncoder {
	return &VideoEncoder{
		diffEncoder: NewFrameDiffEncoder(iFrameRate, quality),
	}
}

// AddFrame encodes one frame and appends it to the video.
// Automatically decides I-frame vs P-frame.
func (v *VideoEncoder) AddFrame(img image.Image) {
	frame := v.diffEncoder.Feed(img)
	v.Frames = append(v.Frames, frame)
}

// Stats returns compression info for logging/response
func (v *VideoEncoder) Stats() map[string]int {
	iFrames, pFrames, skippedBlocks, encodedBlocks := 0, 0, 0, 0

	for _, f := range v.Frames {
		if f.Type == IFrame {
			iFrames++
		} else {
			pFrames++
			for _, row := range f.Diff {
				for _, b := range row {
					if b.Skip {
						skippedBlocks++
					} else {
						encodedBlocks++
					}
				}
			}
		}
	}

	return map[string]int{
		"i_frames":       iFrames,
		"p_frames":       pFrames,
		"skipped_blocks": skippedBlocks,
		"encoded_blocks": encodedBlocks,
	}
}
