package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"strconv"
	"sync"
)

const numWorkers = 4 // Número de trabajadores en el pool

func main() {
	archivo, err := os.Open("measurements_50M.txt")
	if err != nil {
		fmt.Println("Error al abrir el archivo:", err)
		return
	}
	defer archivo.Close()

	reader := bufio.NewReader(archivo)
	estaciónTemperatura := make(map[string][]float64)
	var wg sync.WaitGroup

	// Crear canales para comunicarse entre el main y los trabajadores
	lines := make(chan []byte, numWorkers)
	results := make(chan map[string][]float64, numWorkers)

	// Lanzar trabajadores
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go worker(lines, results, &wg)
	}

	// Enviar líneas a los trabajadores
	go func() {
		defer close(lines)
		for {
			linea, err := reader.ReadBytes('\n')
			if err != nil {
				break
			}
			lines <- linea
		}
	}()

	// Cerrar results cuando todos los trabajadores hayan terminado
	go func() {
		wg.Wait()
		close(results)
	}()

	// Recopilar resultados de los trabajadores
	for result := range results {
		for key, value := range result {
			if array, exist := estaciónTemperatura[key]; !exist {
				estaciónTemperatura[key] = value
			} else {
				array[1] += value[1]
				array[3] += value[3]
				if value[0] < array[0] {
					array[0] = value[0]
				}
				if value[2] > array[2] {
					array[2] = value[2]
				}
			}
		}
	}

	// Resto del código para ordenar y mostrar resultados
	var claves [][]byte
	for clave := range estaciónTemperatura {
		claves = append(claves, []byte(clave))
	}
	SortBytes(claves)

	fmt.Print("{")
	for i, clave := range claves {
		array := estaciónTemperatura[string(clave)]
		min := Round(array[0])
		average := Round(array[1] / array[3])
		max := Round(array[2])

		fmt.Printf("%s=%.1f/%.1f/%.1f", clave, min, average, max)

		if i+1 == len(claves) {
			break
		}
		fmt.Print(", ")
	}
	fmt.Print("}")
}

func worker(lines <-chan []byte, results chan<- map[string][]float64, wg *sync.WaitGroup) {
	defer wg.Done()

	workerResult := make(map[string][]float64)

	for linea := range lines {
		semicolonIndex := -1
		for i, char := range linea {
			if char == ';' {
				semicolonIndex = i
				break
			}
		}

		nombre := linea[:semicolonIndex]
		numero, err := strconv.ParseFloat(string(linea[semicolonIndex+1:len(linea)-1]), 64)
		if err != nil {
			fmt.Println("Error al convertir la temperatura a número:", err)
			return
		}

		if array, exist := workerResult[string(nombre)]; !exist {
			workerResult[string(nombre)] = []float64{numero, numero, numero, 1.0}
		} else {
			array[1] += numero
			array[3]++
			if numero < array[0] {
				array[0] = numero
			}
			if numero > array[2] {
				array[2] = numero
			}
		}
	}

	results <- workerResult
}

func Round(num float64) float64 {
	if num >= 0 {
		return math.Round(num*10.0) / 10.0
	}
	return math.RoundToEven(num*10.0) / 10.0
}

// SortBytes ordena un slice de []byte alfabéticamente.
func SortBytes(slices [][]byte) {
	for i := 0; i < len(slices)-1; i++ {
		for j := i + 1; j < len(slices); j++ {
			if string(slices[i]) > string(slices[j]) {
				slices[i], slices[j] = slices[j], slices[i]
			}
		}
	}
}
