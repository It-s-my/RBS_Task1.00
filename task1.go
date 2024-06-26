package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

// функция, printUsage вызывает функцию PrintDefaults, которая выводит список опций -h, -help для помощи вывода в консоль
func printUsage() {
	fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
	fmt.Fprintf(flag.CommandLine.Output(), "  %s -src <source_file> -dst <destination_directory>\n", os.Args[0])
	flag.PrintDefaults()
}

//функция обрабатывает ссылки из файла и отправляет запросы, результат записывает в отдельный созданный файл
func prUrl(url *url.URL, outputFileName string, wg *sync.WaitGroup) {
	defer wg.Done()
	err := processURL(url, outputFileName, wg)
	if err != nil {
		fmt.Println("Error", err)
		return
	}

}
func processURL(url *url.URL, outputFileName string, wg *sync.WaitGroup) error {
	// Проверяем, существует ли уже файл с таким именем
	if _, err := os.Stat(outputFileName); err == nil {
		fmt.Println("Файл для", url.Host, "уже существует")
		return err
	}

	resp, err := http.Get(url.String())
	if err != nil {
		fmt.Println(url, ":", err)
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(url, ":", err)
		return err
	}

	outputFile, err := os.Create(outputFileName)
	if err != nil {
		fmt.Println(url, ":", err)
		return err
	}
	defer outputFile.Close()

	outputFile.Write(body)

	fmt.Println(url, ":", "Результат сохранен в файл", outputFileName)
	return nil
}

func createOutputFileName(url *url.URL, dstPtr string) {
	fmt.Sprintf("%s/%s.txt", dstPtr, url.Host)
}

func main() {
	var wg sync.WaitGroup
	start := time.Now()                             //счетчик выполнения программы
	fileName := flag.String("src", "", "Имя файла") //объявляем флаги
	dstPtr := flag.String("dst", "", "Название конечной директории")
	flag.Parse()

	if *fileName == "" {
		fmt.Println("Необходимо указать имя файла")
		printUsage()
		return
	}

	if *dstPtr == "" {
		fmt.Println("Необходимо указать название конечной директории")
		printUsage()
		return
	}

	file, err := os.ReadFile(*fileName)
	if err != nil {
		fmt.Println("Ошибка чтения файла:", err)
		return
	}

	if _, err := os.Stat(*dstPtr); err == nil {
		if fileInfo, _ := os.Stat(*dstPtr); fileInfo.IsDir() {
			fmt.Println("Директория", *dstPtr, "уже создана")
		} else {
			fmt.Println(*dstPtr, "не является директорией")
			return
		}
	} else if os.IsNotExist(err) {
		fmt.Println("Директория", *dstPtr, "не существует. Создаем...")
		err = os.MkdirAll(*dstPtr, 0777)
		if err != nil {
			fmt.Println("Ошибка создания директории:", err)
			return
		}
		fmt.Println("Директория", *dstPtr, "успешно создана")
	} else {
		fmt.Println("Ошибка при проверке директории:", err)
		return
	}

	content := string(file)
	lines := strings.Split(content, "\n")
	//парсинг ссылок из файла и вызов функции processURL
	for _, line := range lines {
		u, err := url.Parse(line)
		if err == nil && u.Scheme != "" && u.Host != "" {
			wg.Add(1)
			go prUrl(u, *dstPtr, &wg)
			createOutputFileName(u, *dstPtr)
		}
	}

	elapsed := time.Since(start) //остановка счётчика и вывод
	wg.Wait()
	fmt.Printf("Время выполнения программы: %s\n", elapsed)
}
