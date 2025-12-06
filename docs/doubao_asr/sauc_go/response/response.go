package response

import (
	"encoding/binary"
	"encoding/json"

	"byted.org/data-speech/asr-tob-demo/sauc/common"
)

type AsrResponsePayload struct {
	AudioInfo struct {
		Duration int `json:"duration"`
	} `json:"audio_info"`
	Result struct {
		Text       string `json:"text"`
		Utterances []struct {
			Definite  bool   `json:"definite"`
			EndTime   int    `json:"end_time"`
			StartTime int    `json:"start_time"`
			Text      string `json:"text"`
			Words     []struct {
				EndTime   int    `json:"end_time"`
				StartTime int    `json:"start_time"`
				Text      string `json:"text"`
			} `json:"words"`
		} `json:"utterances,omitempty"`
	} `json:"result"`
	Error string `json:"error,omitempty"`
}

type AsrResponse struct {
	Code            int                 `json:"code"`
	Event           int                 `json:"event"`
	IsLastPackage   bool                `json:"is_last_package"`
	PayloadSequence int32               `json:"payload_sequence"`
	PayloadSize     int                 `json:"payload_size"`
	PayloadMsg      *AsrResponsePayload `json:"payload_msg"`
}

func ParseResponse(msg []byte) *AsrResponse {
	var result AsrResponse

	headerSize := msg[0] & 0x0f
	messageType := common.MessageType(msg[1] >> 4)
	messageTypeSpecificFlags := common.MessageTypeSpecificFlags(msg[1] & 0x0f)
	serializationMethod := common.SerializationType(msg[2] >> 4)
	messageCompression := common.CompressionType(msg[2] & 0x0f)
	payload := msg[headerSize*4:]
	// 解析messageTypeSpecificFlags
	if messageTypeSpecificFlags&0x01 != 0 {
		result.PayloadSequence = int32(binary.BigEndian.Uint32(payload[:4]))
		payload = payload[4:]
	}
	if messageTypeSpecificFlags&0x02 != 0 {
		result.IsLastPackage = true
	}
	if messageTypeSpecificFlags&0x04 != 0 {
		result.Event = int(binary.BigEndian.Uint32(payload[:4]))
		payload = payload[4:]
	}

	// 解析messageType
	switch messageType {
	case common.SERVER_FULL_RESPONSE:
		result.PayloadSize = int(binary.BigEndian.Uint32(payload[:4]))
		payload = payload[4:]
	case common.SERVER_ERROR_RESPONSE:
		result.Code = int(binary.BigEndian.Uint32(payload[:4]))
		result.PayloadSize = int(binary.BigEndian.Uint32(payload[4:8]))
		payload = payload[8:]
	}

	if len(payload) == 0 {
		return &result
	}

	// 是否压缩
	if messageCompression == common.GZIP {
		payload = common.GzipDecompress(payload)
	}

	// 解析payload
	var asrResponse AsrResponsePayload
	switch serializationMethod {
	case common.JSON:
		_ = json.Unmarshal(payload, &asrResponse)
	case common.NO_SERIALIZATION:
	}
	result.PayloadMsg = &asrResponse
	return &result
}
