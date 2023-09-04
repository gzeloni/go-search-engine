package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("Digite o termo de busca (ou 'sair' para encerrar): ")
	}
}
