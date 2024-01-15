package main

import (
	"cloud.google.com/go/speech/apiv1"
	"context"
	"fmt"
	"github.com/go-audio/audio"
	"github.com/go-audio/audio/wav"
	speechpb "google.golang.org/genproto/googleapis/cloud/speech/v1"
	"log"
	"os"
	"os/signal"
)

func main() {
	ctx := context.Background()

	// Inicializa o cliente da API de reconhecimento de fala
	client, err := speech.NewClient(ctx)
	if err != nil {
		log.Fatalf("Erro ao criar cliente de reconhecimento de fala: %v", err)
	}

	// Configuração do formato do arquivo de áudio
	formato := &audio.Format{SampleRate: 16000, NumChannels: 1}

	// Abre um arquivo WAV para gravar
	arquivoWAV, err := os.Create("gravacao.wav")
	if err != nil {
		log.Fatalf("Erro ao criar arquivo WAV: %v", err)
	}
	defer arquivoWAV.Close()

	gravador := wav.NewEncoder(arquivoWAV, formato.SampleRate, 16, 1, 1)

	// Canal para receber os sinais de interrupção (Ctrl+C)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	fmt.Println("Pressione Ctrl+C para encerrar a gravação.")

	// Inicia a gravação em uma goroutine
	go func() {
		defer gravador.Close()

		// Configuração do buffer de áudio
		bufferSize := formato.SampleRate * formato.NumChannels
		buf := &audio.IntBuffer{Data: make([]int, bufferSize)}

		for {
			select {
			case <-interrupt:
				return // Encerra a gravação quando o sinal de interrupção é recebido
			default:
				// Lê o áudio do microfone
				if err := audio.Record(buf, formato); err != nil {
					log.Fatalf("Erro ao gravar áudio: %v", err)
				}

				// Grava o áudio no arquivo WAV
				if err := gravador.Write(buf); err != nil {
					log.Fatalf("Erro ao gravar áudio no arquivo WAV: %v", err)
				}
			}
		}
	}()

	// Aguarda até o sinal de interrupção ser recebido
	<-interrupt

	// Fecha o cliente da API de reconhecimento de fala
	if err := client.Close(); err != nil {
		log.Fatalf("Erro ao fechar cliente de reconhecimento de fala: %v", err)
	}

	// Chama a função para transcrever o áudio
	transcreverAudio(ctx, "gravacao.wav", client)
}

func transcreverAudio(ctx context.Context, arquivo string, client *speech.Client) {
	// Abre o arquivo de áudio para leitura
	arquivoAudio, err := os.Open(arquivo)
	if err != nil {
		log.Fatalf("Erro ao abrir arquivo de áudio: %v", err)
	}
	defer arquivoAudio.Close()

	// Configuração da solicitação de transcrição
	configuracao := &speechpb.RecognitionConfig{
		Encoding:        speechpb.RecognitionConfig_LINEAR16,
		SampleRateHertz: 16000,
		LanguageCode:    "pt-BR",
	}

	req := &speechpb.RecognizeRequest{
		Config: configuracao,
		Audio: &speechpb.RecognitionAudio{
			AudioSource: &speechpb.RecognitionAudio_Content{Content: lerAudio(arquivoAudio)},
		},
	}

	// Envia a solicitação de transcrição
	resp, err := client.Recognize(ctx, req)
	if err != nil {
		log.Fatalf("Erro ao chamar a API de reconhecimento de fala: %v", err)
	}

	// Exibe os resultados
	for _, resultado := range resp.Results {
		for _, alternativa := range resultado.Alternatives {
			fmt.Printf("Texto transcrito: %s\n", alternativa.Transcript)
		}
	}
}

// Função auxiliar para ler o conteúdo do arquivo de áudio
func lerAudio(arquivo *os.File) []byte {
	info, _ := arquivo.Stat()
	tamanho := info.Size()
	bytes := make([]byte, tamanho)

	arquivo.Read(bytes)

	return bytes
}
