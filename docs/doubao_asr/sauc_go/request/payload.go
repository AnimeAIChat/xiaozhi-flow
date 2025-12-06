package request

import (
	"bytes"
	"encoding/binary"

	"github.com/bytedance/sonic"

	"byted.org/data-speech/asr-tob-demo/sauc/common"
)

type UserMeta struct {
	Uid        string `json:"uid,omitempty"`
	Did        string `json:"did,omitempty"`
	Platform   string `json:"platform,omitempty" `
	SDKVersion string `json:"sdk_version,omitempty"`
	APPVersion string `json:"app_version,omitempty"`
}

type AudioMeta struct {
	Format  string `json:"format,omitempty"`
	Codec   string `json:"codec,omitempty"`
	Rate    int    `json:"rate,omitempty"`
	Bits    int    `json:"bits,omitempty"`
	Channel int    `json:"channel,omitempty"`
}

type CorpusMeta struct {
	BoostingTableName string `json:"boosting_table_name,omitempty"`
	CorrectTableName  string `json:"correct_table_name,omitempty"`
	Context           string `json:"context,omitempty"`
}

type RequestMeta struct {
	ModelName       string     `json:"model_name,omitempty"`
	EnableITN       bool       `json:"enable_itn,omitempty"`
	EnablePUNC      bool       `json:"enable_punc,omitempty"`
	EnableDDC       bool       `json:"enable_ddc,omitempty"`
	ShowUtterances  bool       `json:"show_utterances"`
	EnableNonstream bool       `json:"enable_nonstream"`
	Corpus          CorpusMeta `json:"corpus,omitempty"`
}

type AsrRequestPayload struct {
	User    UserMeta    `json:"user"`
	Audio   AudioMeta   `json:"audio"`
	Request RequestMeta `json:"request"`
}

func NewFullClientRequest() []byte {
	var request bytes.Buffer
	request.Write(DefaultHeader().WithMessageTypeSpecificFlags(common.POS_SEQUENCE).toBytes())
	payload := AsrRequestPayload{
		User: UserMeta{
			Uid: "demo_uid",
		},
		Audio: AudioMeta{
			Format:  "wav",
			Codec:   "raw",
			Rate:    16000,
			Bits:    16,
			Channel: 1,
		},
		Request: RequestMeta{
			ModelName:       "bigmodel",
			EnableITN:       true,
			EnablePUNC:      true,
			EnableDDC:       true,
			ShowUtterances:  true,
			EnableNonstream: false,
		},
	}
	payloadArr, _ := sonic.Marshal(payload)
	payloadArr = common.GzipCompress(payloadArr)
	payloadSize := len(payloadArr)
	payloadSizeArr := make([]byte, 4)
	binary.BigEndian.PutUint32(payloadSizeArr, uint32(payloadSize))
	_ = binary.Write(&request, binary.BigEndian, int32(1))
	request.Write(payloadSizeArr)
	request.Write(payloadArr)
	return request.Bytes()
}

func NewAudioOnlyRequest(seq int, segment []byte) []byte {
	var request bytes.Buffer
	header := DefaultHeader()
	if seq < 0 {
		header.WithMessageTypeSpecificFlags(common.NEG_WITH_SEQUENCE)
	} else {
		header.WithMessageTypeSpecificFlags(common.POS_SEQUENCE)
	}
	header.WithMessageType(common.CLIENT_AUDIO_ONLY_REQUEST)
	request.Write(header.toBytes())

	// write seq
	_ = binary.Write(&request, binary.BigEndian, int32(seq))
	// write payload size
	payload := common.GzipCompress(segment)
	_ = binary.Write(&request, binary.BigEndian, int32(len(payload)))
	// write payload
	request.Write(payload)
	return request.Bytes()
}
