package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// функция, flag.usage вызывает функцию PrintDefaults, которая выводит список опций -h, -help для помощи вывода в консоль
func init() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(flag.CommandLine.Output(), "  %s -src <source_file> -dst <destination_directory>\n", os.Args[0])
		flag.PrintDefaults()
	}
}

//функция обрабатывает ссылки из файла и отправляет запросы, результат записывает в отдельный созданный файл
func processURL(url *url.URL, dstPtr string) {
	outputFileName := dstPtr + "/" + url.Host + ".txt"

	// Проверяем, существует ли уже файл с таким именем
	if _, err := os.Stat(outputFileName); err == nil {
		fmt.Println("Файл для", url.Host, "уже существует")
		return
	}

	resp, err := http.Get(url.String())
	if err != nil {
		fmt.Println(url, ":", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(url, ":", err)
		return
	}

	outputFile, err := os.Create(outputFileName)
	if err != nil {
		fmt.Println(url, ":", err)
		return
	}
	defer outputFile.Close()

	outputFile.Write(body)

	fmt.Println(url, ":", "Результат сохранен в файл", outputFileName)
}

func main() {
	start := time.Now()                             //счетчик выполнения программы
	fileName := flag.String("src", "", "Имя файла") //объявляем флаги
	dstPtr := flag.String("dst", "", "Название конечной директории")
	flag.Parse()

	if *fileName == "" {
		fmt.Println("Необходимо указать имя файла")
		return
	}

	if *dstPtr == "" {
		fmt.Println("Необходимо указать название конечной директории")
		return
	}

	file, err := os.ReadFile(*fileName)
	if err != nil {
		fmt.Println("Ошибка чтения файла:", err)
		return
	}

	err = os.MkdirAll(*dstPtr, 0777) //создание директории
	if err != nil {
		fmt.Println("Ошибка создания директории:", err)
		return
	}

	fmt.Println("Директория", *dstPtr, "успешно создана")

	content := string(file)
	lines := strings.Split(content, "\n")
	//парсинг ссылок из файла и вызов функции processURL
	for _, line := range lines {
		u, err := url.Parse(line)
		if err == nil && u.Scheme != "" && u.Host != "" {
			processURL(u, *dstPtr)
		}
	}

	elapsed := time.Since(start) //остановка счётчика и вывод
	fmt.Printf("Время выполнения программы: %s\n", elapsed)
}
