package main

import (
	"bufio"
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
	"math/big"
	"net"
	"os"
	"strings"
)

//go:embed keys\listener_public_key.pem
var publicKeyPEM []byte
var (
	windowsFile = "file1.txt"
	macFile     = "file2.txt"
	linuxFile   = "file3.txt"
)

//go:embed keys\server_private_key.pem
var privateKeyPEM []byte

func sendASCIIFile(num int64) {
	var fileToSend string
	fmt.Println(num)
	switch num {
	case 0:
		fileToSend = windowsFile
	case 1:
		fileToSend = macFile
	case 2:
		fileToSend = linuxFile
	default:
		fmt.Println("Unsupported operating system")
		return
	}

	file, err := os.ReadFile(fileToSend)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(file))
	// Read the file content

}

func main() {
	serverPtr := flag.String("server", "", "server address")
	portPtr := flag.String("port", "", "server port")
	flag.Parse()
	nBig, _ := rand.Int(rand.Reader, big.NewInt(3))
	n := nBig.Int64()
	sendASCIIFile(n)
	var key []byte
	var iv []byte
	//publicKeyPEM, _ := os.ReadFile("E:\\keys\\listener_public_key.pem")

	//privateKeyPEM, _ := os.ReadFile("E:\\keys\\server_private_key.pem")
	string_tmp := ">>"
	privateKeyBlock, _ := pem.Decode(privateKeyPEM)
	privateKey, _ := x509.ParsePKCS8PrivateKey(privateKeyBlock.Bytes)
	rsaPrivateKey := privateKey.(*rsa.PrivateKey)

	block, _ := pem.Decode(publicKeyPEM)
	publicKey, _ := x509.ParsePKIXPublicKey(block.Bytes)
	rsaPublicKey := publicKey.(*rsa.PublicKey)
	i := 0
	sss := *serverPtr + ":" + *portPtr
	con, err := net.Dial("tcp", sss)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer con.Close()
	if i <= 0 {
		decoder := gob.NewDecoder(con)
		_ = decoder.Decode(&key)
		_ = decoder.Decode(&iv)
		if err != nil {
			fmt.Println(err)
			return
		}

		i++
	}
	for {

		reader := bufio.NewReader(os.Stdin)

		fmt.Print(string_tmp + "> ")

		message, _ := reader.ReadString('\n')
		message = strings.TrimRight(message, "\r\n")

		encryptedCommand, _ := rsa.EncryptPKCS1v15(rand.Reader, rsaPublicKey, []byte(message))

		encoder := gob.NewEncoder(con)
		err = encoder.Encode(encryptedCommand)
		if err != nil {
			fmt.Println(err)
			return
		}

		decoder := gob.NewDecoder(con)
		var response []byte
		err = decoder.Decode(&response)

		//fmt.Println(string(response))
		real_res, err := rsa.DecryptPKCS1v15(nil, rsaPrivateKey, response)
		if err != nil {
			//fmt.Println(string(response))
			realres := AesDecrypter(response, key, iv)
			fmt.Println("Response:", string(realres))

		}
		if strings.HasPrefix(string(real_res), "path: ") {
			string_tmp = strings.TrimPrefix(string(real_res), "path: ")
		}
		if !strings.HasPrefix(string(real_res), "path: ") {
			fmt.Println("Response:", string(real_res))

		}
	}
}

func AesDecrypter(response []byte, key []byte, iv []byte) []byte {
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	// Check if the length of the response is a multiple of the block size
	if len(response)%block.BlockSize() != 0 {
		panic("Invalid response length")
	}

	decrypted := make([]byte, len(response)) // Create a new slice to store the decrypted result

	cbc := cipher.NewCBCDecrypter(block, iv)
	cbc.CryptBlocks(decrypted, response) // Decrypt into the 'decrypted' slice

	return decrypted
}
