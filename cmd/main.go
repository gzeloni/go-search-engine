package main

import (
	"bufio"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

func main() {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("Digite o termo de busca (ou 'sair' para encerrar): ")
		searchTerm, _ := reader.ReadString('\n')
		searchTerm = strings.TrimSpace(searchTerm)
		searchTerm = strings.ToLower(searchTerm)

		if searchTerm == "sair" {
			break
		}

		// URLs a serem rastreadas.
		urlsToCrawl := []string{
			"https://g1.globo.com",
			"https://www.cnnbrasil.com.br",
			"https://news.google.com/home?hl=pt-BR&gl=BR&ceid=BR:pt-419",
			"https://www.google.com/search?q=",
		}

		// Percorre as URLs e realiza o rastreamento e busca.
		for _, url := range urlsToCrawl {
			// Condição para substituir espaços por traços apenas no caso do Google.
			if strings.Contains(url, "www.google.com") {
				searchURL := url + strings.ReplaceAll(searchTerm, " ", "+")
				fmt.Println("URL de busca:", searchURL)
				CrawlAndSearch(searchURL, searchTerm)
			} else {
				fmt.Println("URL a ser rastreada:", url)
				CrawlAndSearch(url, searchTerm)
			}
		}
	}
}

// CrawlAndSearch inicia o rastreamento e busca de termos em uma URL.
func CrawlAndSearch(url string, searchTerm string) {
	crawledURLs := make(map[string]bool)
	RecursiveCrawl(url, searchTerm, crawledURLs)
}

// ExtractLinksFromNode extrai todos os links de um nó HTML.
func ExtractLinksFromNode(node *html.Node) []string {
	var links []string

	// Função interna recursiva para extrair links de nós HTML.
	var extractLinks func(*html.Node)
	extractLinks = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key == "href" {
					links = append(links, attr.Val)
					break
				}
			}
		}
		// Recursivamente chama a função para os filhos do nó atual.
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			extractLinks(child)
		}
	}

	extractLinks(node)
	return links
}

func FindNode(parentNode *html.Node) *html.Node {
	validNodeNames := []string{"span", "h1", "h2", "h3", "h4"}

	for child := parentNode.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == html.ElementNode {
			for _, nodeName := range validNodeNames {
				if child.Data == nodeName {
					return child
				}
			}
		}
	}
	return nil
}

func GetTitleFromLink(linkNode *html.Node) string {
	var title string

	doc := goquery.NewDocumentFromNode(linkNode)

	doc.Find("span, h1, h2, h3, h4").Each(func(i int, s *goquery.Selection) {
		title = strings.TrimSpace(s.Text())
	})

	return title
}

// RecursiveCrawl realiza o rastreamento recursivo de URLs.
func RecursiveCrawl(url string, searchTerm string, crawledURLs map[string]bool) {
	// Verifica se atingiu o limite de URLs rastreadas.
	if len(crawledURLs) >= 100 {
		return
	}
	// Verifica se a URL já foi rastreada anteriormente.
	if crawledURLs[url] {
		return
	}
	// Marca a URL atual como rastreada.
	crawledURLs[url] = true

	// Realiza uma requisição HTTP GET para a URL.
	response, err := http.Get(url)
	if err != nil {
		fmt.Printf("Erro ao fazer a requisição para %s: %v\n", url, err)
		return
	}
	defer response.Body.Close()

	// Verifica se o status da resposta é OK.
	if response.StatusCode != http.StatusOK {
		fmt.Printf("Erro de status para %s: %s\n", url, response.Status)
		return
	}

	// Realiza o parsing do HTML da resposta.
	htmlDocument, err := html.Parse(response.Body)
	if err != nil {
		fmt.Printf("Erro ao fazer o parsing do HTML para %s: %v\n", url, err)
		return
	}

	var lastTitle, lastLink string
	// Inicia o processo de busca e rastreamento.
	Spider(htmlDocument, searchTerm, crawledURLs, &lastTitle, &lastLink)
}

// Spider realiza a busca de termos e rastreamento de URLs em um nó HTML.
func Spider(node *html.Node, searchTerm string, crawledURLs map[string]bool, lastTitle *string, lastLink *string) {
	if node.Type == html.ElementNode && node.Data == "a" {
		// Inicializa link como uma string vazia.
		link := ""

		// Procura pelo atributo "href".
		for _, attr := range node.Attr {
			if attr.Key == "href" {
				link = attr.Val
				break
			}
		}

		// Verifica se o link não está vazio e é um link HTTP/HTTPS válido.
		if link != "" {
			parsedURL, err := url.Parse(link)
			if err == nil && (parsedURL.Scheme == "http" || parsedURL.Scheme == "https") {
				// Verifica se o link contém o termo de busca.
				if containsAllWords(link, searchTerm) {
					// fmt.Println("Comparando link:", link, "com searchTerm:", searchTerm)
					title := GetTitleFromLink(node)
					// Verifica se o título é válido e se é diferente do título anterior.
					if title != "" && (title != *lastTitle || link != *lastLink) {
						fmt.Println("Título encontrado:", title)
						fmt.Println("Link encontrado:", link)
						*lastTitle = title
						*lastLink = link
					}
				}

				// Verifica se o link ainda não foi rastreado.
				if !crawledURLs[link] {
					crawledURLs[link] = true // Marca o link como rastreado
					// RecursiveCrawl(link, searchTerm, crawledURLs) // Você pode adicionar essa chamada se necessário
				}
			}
		}
	}

	// Recursivamente chama a função para os filhos do nó atual.
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		Spider(child, searchTerm, crawledURLs, lastTitle, lastLink)
	}
}

// containsAllWords verifica se uma string contém todas as palavras de outra string.
func containsAllWords(s, words string) bool {
	wordList := strings.Fields(words)
	missingWords := []string{} // Para armazenar as palavras ausentes

	for _, word := range wordList {
		if !strings.Contains(s, word) {
			missingWords = append(missingWords, word)
		}
	}

	if len(missingWords) > 0 {
		return false
	}

	return true
}
