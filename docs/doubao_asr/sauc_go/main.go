package main

import (
	"context"
	"flag"
	"log"
	"os"
	"sync"

	"github.com/bytedance/sonic"

	"byted.org/data-speech/asr-tob-demo/sauc/client"
	"byted.org/data-speech/asr-tob-demo/sauc/response"
)

// /Users/bytedance/Downloads/flow_20230824_e4729a02-7fcf-48e3-8e39-0d1ba61f8e86.wav

var filePath = flag.String("file", "/Users/bytedance/code/python/eng_ddc_itn.wav", "audio file path")

//var filePath = flag.String("file", "/Users/bytedance/Downloads/chinese_ddc_itn (1).wav", "audio file path")

//var wsURL = flag.String("url", "wss://speech-lf.byted.org/api/v3/sauc_test/v2", "request url")

// var wsURL = flag.String("url", "wss://speech-hl.byted.org/api/v3/press_test/sauc/bigmodel_nostream", "request url")
//var wsURL = flag.String("url", "wss://openspeech.byted.org/api/v3/sauc/bigmodel_nostream", "request url")

var wsURL = flag.String("url", "wss://openspeech.bytedance.com/api/v3/sauc/bigmodel_nostream", "request url")

// var wsURL = flag.String("url", "wss://openspeech.bytedance.com/api/v3/sauc/bigmodel_async", "request url")
//var wsURL = flag.String("url", "wss://openspeech.bytedance.com/api/v3/sauc/bigmodel", "request url")

// var wsURL = flag.String("url", "wss://openspeech.bytedance.com/api/v3/test/sauc/v3/bigmodel", "request url")
var segmentDuration = flag.Int("seg_duration", 200, "audio duration(ms) per packet, default:100")

func main() {
	flag.Parse()

	// 打开日志文件，如果文件不存在则创建
	file, err := os.OpenFile("run.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND|os.O_TRUNC, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer file.Close()
	log.SetFlags(log.Lmicroseconds | log.Lshortfile)
	do()
}

func do() {
	c := client.NewAsrWsClient(*wsURL, *segmentDuration)
	resChan := make(chan *response.AsrResponse)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		for res := range resChan {
			resStr, _ := sonic.MarshalString(res)
			log.Println(resStr)
		}
		wg.Done()
	}()

	err := c.Excute(context.Background(), *filePath, resChan)
	if err != nil {
		log.Fatalf("failed to excute: %v", err)
		return
	}
	wg.Wait()
}
