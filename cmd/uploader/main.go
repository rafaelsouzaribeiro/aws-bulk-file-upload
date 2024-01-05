package main

import (
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/rafaelsouzaribeiro/aws-bulk-file-upload/configs"
)

var (
	s3Client *s3.S3
	s3Bucket string
	wg       sync.WaitGroup
)

func init() {
	env, err := configs.LoadConfig("./")
	if err != nil {
		panic(err)
	}

	sess, err := session.NewSession(
		&aws.Config{
			Region: aws.String("us-east-1"),
			Credentials: credentials.NewStaticCredentials(
				env.AcessKey,
				env.SecretKey,
				"",
			),
		},
	)

	if err != nil {
		panic(err)
	}

	s3Client = s3.New(sess)
	s3Bucket = env.Bucket
}

func main() {
	dir, err := os.Open("./tmp")

	if err != nil {
		panic(err)
	}

	defer dir.Close()

	// Máximo de 100 channel quando encher só fazer upload depois de esvaziar o channel
	// usou struct pq é o menor valor em memória
	uploadChannel := make(chan struct{}, 100)

	// Retenta 10 vezes se houver erro no canal
	errorFileUpload := make(chan string, 10)

	go func() {
		for {
			select {
			case filename := <-errorFileUpload:
				uploadChannel <- struct{}{}
				wg.Add(1)
				go UploadFile(filename, uploadChannel, errorFileUpload)

			}
		}
	}()

	for {
		files, err := dir.ReadDir(1)
		if err != nil {
			if err == io.EOF {
				break
			}

			fmt.Printf("Error ao ler o diretório: %s\n", err)
			continue
		}

		wg.Add(1)
		// Máximo de 100 channel quando encher só fazer upload depois de esvaziar o channel
		// usou struct pq é o menor valor em memória
		uploadChannel <- struct{}{}
		go UploadFile(files[0].Name(), uploadChannel, errorFileUpload)

	}

	wg.Wait()
}

// somente leitura <-
func UploadFile(fileName string, uploadChannel <-chan struct{}, errorFileUpload chan<- string) {
	defer wg.Done()
	completeFilename := fmt.Sprintf("./tmp/%s", fileName)

	fmt.Printf("Upload do arquivo %s para buker %s iniciou\n", completeFilename, s3Bucket)

	f, err := os.Open(completeFilename)
	if err != nil {
		fmt.Printf("Error ao abrir arquivo tal: %s\n", completeFilename)
		<-uploadChannel // esvazia o canal
		errorFileUpload <- completeFilename
		return
	}
	defer f.Close()
	_, err = s3Client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(s3Bucket),
		Key:    aws.String(fileName),
		Body:   f,
	})

	if err != nil {
		fmt.Printf("Error ao fazer upload do arquivo tal: %s\n", completeFilename)
		<-uploadChannel // esvazia o canal
		errorFileUpload <- completeFilename
		return
	}

	fmt.Printf("Arquivo %s upload com sucesso\n", completeFilename)
	<-uploadChannel // esvazia o canal
}
