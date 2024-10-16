package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"syscall"
	"time"
)

var monthNames = map[string]string{
	"January":   "janeiro",
	"February":  "fevereiro",
	"March":     "marco",
	"April":     "abril",
	"May":       "maio",
	"June":      "junho",
	"July":      "julho",
	"August":    "agosto",
	"September": "setembro",
	"October":   "outubro",
	"November":  "novembro",
	"December":  "dezembro",
}

var monthNumber = map[string]int{
	"January":   1,
	"February":  2,
	"March":     3,
	"April":     4,
	"May":       5,
	"June":      6,
	"July":      7,
	"August":    8,
	"September": 9,
	"October":   10,
	"November":  11,
	"December":  12,
}

var folderName string

var rgx *regexp.Regexp

func main() {
	//
	// 1. Iterar sobre cada arquivo
	// 		1.1. Verificar a data de criação de arquivo
	//		1.2. Se não existir uma pasta com o mês de criação
	//				1.2.1. Crie uma pasta para esse mês
	//		1.3. Mover o arquivo para a pasta de seu mês
	//

	// Definir a regex para o formato "yyyymmdd_id.ext"
	pattern := `^(\d{4})(\d{2})(\d{2})_(\w+)\.(\w+)$`
	rgx = regexp.MustCompile(pattern)

	start := time.Now()

	args := os.Args[1:]
	if len(args) > 0 {
		folderName = args[0]
	} else {
		folderName = "."
	}
	entries, err := os.ReadDir(folderName)
	if err != nil {
		fmt.Println("erro ao abrir a pasta! erro: ", err.Error())
		os.Exit(1)
	}

	totalFiles := 0

	for _, e := range entries {
		if !e.IsDir() {
			var monthFolder string

			creation := getCreationDate(e)

			if !existsMonthFolder(creation) {
				monthFolder, err = createMonthFolder(creation)
				if err != nil {
					year, month, _ := creation.Date()
					log.Printf("Não consigo criar pasta para mês %v-%v\n", month, year)
				}
			} else {
				year, month, _ := creation.Date()
				monthFolder = fmt.Sprintf("%v-%v", monthNames[month.String()], year)
			}

			err := moveFile(e, monthFolder)
			if err == nil {
				totalFiles++
			}
		}
	}

	duration := time.Since(start)

	fmt.Printf("Total de arquivos movidos: %v arquivos movidos\n", totalFiles)
	fmt.Printf("Tempo passado:             %v milliseconds\n", duration.Milliseconds())
}

func getCreationDate(e os.DirEntry) time.Time {
	var date time.Time
	dateByName, err := getDateFromFileName(e.Name())
	if err == nil {
		date = dateByName
	} else {
		date = getByDateByModTime(e)
	}
	return date
}

// get the creation date by the name of the file
func getDateFromFileName(fileName string) (time.Time, error) {
	// Verificar se o nome do arquivo corresponde ao padrão
	match := rgx.FindStringSubmatch(fileName)

	if len(match) == 0 {
		return time.Time{}, fmt.Errorf("arquivo não corresponde ao padrão")
	}

	// Capturar os grupos de ano, mês e dia
	year, err := strconv.Atoi(match[1])
	if err != nil {
		return time.Time{}, err
	}
	month, err := strconv.Atoi(match[2])
	if err != nil {
		return time.Time{}, err
	}
	day, err := strconv.Atoi(match[3])
	if err != nil {
		return time.Time{}, err
	}

	// Construir o objeto time.Time
	date := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)

	if date.After(time.Now()) {
		return time.Time{}, fmt.Errorf("data extraída é posterior à data atual")
	}

	return date, nil
}

func getByDateByModTime(e os.DirEntry) time.Time {
	info, _ := e.Info()
	d := info.Sys().(*syscall.Win32FileAttributeData)
	return time.Unix(0, d.CreationTime.Nanoseconds())
}

func createMonthFolder(creation time.Time) (string, error) {
	year, month, _ := creation.Date()
	folderName := fmt.Sprintf("%v-%v", monthNames[month.String()], year)
	err := os.Mkdir(folderName, 0777)
	return folderName, err
}

func existsMonthFolder(creation time.Time) bool {
	year, month, _ := creation.Date()
	filename := fmt.Sprintf("%v-%v", monthNames[month.String()], year)
	_, err := os.Stat(filename)
	if err == nil {
		return true // O arquivo existe
	}
	if os.IsNotExist(err) {
		return false // O arquivo não existe
	}
	return false // Outro erro, mas tratamos como não existente
}

func moveFile(e os.DirEntry, newPath string) error {
	// dir, _ := os.Getwd()
	oldPath := filepath.Join(folderName, e.Name())
	err := os.Rename(oldPath, filepath.Join(newPath, e.Name()))
	if err != nil {
		fmt.Println("Erro ao mover o arquivo:", err)
		return err
	}
	return nil
}
