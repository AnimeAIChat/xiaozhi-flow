package components

import (
	"xiaozhi-server-go/internal/platform/logging"
	"xiaozhi-server-go/internal/utils"
)

// AudioProcessor handles audio data processing
type AudioProcessor struct {
	logger      *logging.Logger
	format      string
	opusDecoder *utils.OpusDecoder
}

// NewAudioProcessor creates a new AudioProcessor
func NewAudioProcessor(logger *logging.Logger, format string) *AudioProcessor {
	ap := &AudioProcessor{
		logger: logger,
		format: format,
	}

	if format == "opus" {
		decoder, err := utils.NewOpusDecoder(&utils.OpusDecoderConfig{
			SampleRate:  16000,
			MaxChannels: 1,
		})
		if err != nil {
			logger.Error("Failed to initialize Opus decoder: %v", err)
		} else {
			ap.opusDecoder = decoder
		}
	}

	return ap
}

// ProcessAudio processes incoming audio data (e.g., decoding)
func (ap *AudioProcessor) ProcessAudio(data []byte) ([]byte, error) {
	if ap.format == "pcm" {
		return data, nil
	} else if ap.format == "opus" {
		if ap.opusDecoder != nil {
			decodedData, err := ap.opusDecoder.Decode(data)
			if err != nil {
				ap.logger.Error("Failed to decode Opus audio: %v", err)
				// Return original data on failure as fallback, or error?
				// Original logic returned original data to queue
				return data, nil
			}
			ap.logger.Debug("Opus decoded: %d bytes -> %d bytes", len(data), len(decodedData))
			return decodedData, nil
		}
		// No decoder, return raw
		return data, nil
	}
	return data, nil
}

// UpdateFormat updates the audio format and re-initializes decoder if needed
func (ap *AudioProcessor) UpdateFormat(format string, sampleRate int, channels int) {
	ap.format = format
	if format == "opus" {
		decoder, err := utils.NewOpusDecoder(&utils.OpusDecoderConfig{
			SampleRate:  sampleRate,
			MaxChannels: channels,
		})
		if err != nil {
			ap.logger.Error("Failed to re-initialize Opus decoder: %v", err)
			ap.opusDecoder = nil
		} else {
			ap.opusDecoder = decoder
		}
	} else {
		ap.opusDecoder = nil
	}
}
