package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/dmulholland/scribble"
	"github.com/go-audio/audio"
	"github.com/go-audio/audio/wav"
	speechpb "google.golang.org/genproto/googleapis/cloud/speech/v1"
)

func main() {
	recorder, err := audio.NewRecorder()
	if err != nil {
		log.Fatal(err)
	}
	defer recorder.Close()

	db, err := scribble.New("./mytalk", nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Inicie a fala para começar a transcrição...")

	for {
		buf, err := recorder.Read()
		if err != nil {
			log.Fatal(err)
		}

		saveWAV(buf)

		text, err := transcreverWAV("gravacao.wav")
		if err != nil {
			log.Println("Erro ao transcrever áudio:", err)
			continue
		}

		err = db.Write("mytalk", nil, []byte(text))
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Texto transcrito:", text)
	}
}

func saveWAV(buf []int) {
	file, err := os.Create("gravacao.wav")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	enc := wav.NewEncoder(file, audio.Int, 1, 44100, 16)
	defer enc.Close()

	if err := enc.Write(&audio.IntBuffer{Data: buf}); err != nil {
		log.Fatal(err)
	}
}

func transcreverWAV(arquivo string) (string, error) {
	ctx := context.Background()

	client, err := speech.NewClient(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	file, err := os.Open(arquivo)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	resp, err := client.Recognize(ctx, &speechpb.RecognizeRequest{
		Config: &speechpb.RecognitionConfig{
			Encoding:        speechpb.RecognitionConfig_LINEAR16,
			SampleRateHertz: 44100,
			LanguageCode:    "en-US",
		},
		Audio: &speechpb.RecognitionAudio{
			AudioSource: &speechpb.RecognitionAudio_Content{Content: lerWAV(file)},
		},
	})
	if err != nil {
		return "", err
	}

	var texto string
	for _, result := range resp.Results {
		texto += result.Alternatives[0].Transcript
	}

	return texto, nil
}

func lerWAV(file *os.File) []byte {
	wavReader, err := wav.NewDecoder(file)
	if err != nil {
		log.Fatal(err)
	}
	defer wavReader.Close()

	buf, err := wavReader.ReadFull()
	if err != nil {
		log.Fatal(err)
	}

	return buf
}
