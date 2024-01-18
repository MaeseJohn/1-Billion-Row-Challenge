package main

import (
	"bufio"
	"os"
	"testing"
)

func compararArchivos(archivo1, archivo2 string) (sonIguales bool, err error) {
	// Abrir el primer archivo
	file1, err := os.Open(archivo1)
	if err != nil {
		return false, err
	}
	defer file1.Close()

	// Abrir el segundo archivo
	file2, err := os.Open(archivo2)
	if err != nil {
		return false, err
	}
	defer file2.Close()

	// Comparar línea por línea
	scanner1 := bufio.NewScanner(file1)
	scanner2 := bufio.NewScanner(file2)

	for scanner1.Scan() && scanner2.Scan() {
		linea1 := scanner1.Text()
		linea2 := scanner2.Text()

		if linea1 != linea2 {
			return false, nil
		}
	}

	// Verificar si ambos archivos tienen la misma cantidad de líneas
	if scanner1.Scan() != scanner2.Scan() {
		return false, nil
	}

	// No hubo diferencias encontradas
	return true, nil
}

func TestCompararArchivos(t *testing.T) {
	archivo1 := "expectedResults_1k.txt"
	archivo2 := "myResult_1k.txt"

	iguales, err := compararArchivos(archivo1, archivo2)
	if err != nil {
		t.Errorf("Error al comparar archivos: %v", err)
	}

	if !iguales {
		t.Errorf("Los archivos no son iguales.")
	}
}
