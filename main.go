package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
)

func main() {
	archivo, err := os.Open("measurements_50M.txt")
	if err != nil {
		fmt.Println("Error al abrir el archivo:", err)
		return
	}
	defer archivo.Close()

	reader := bufio.NewReader(archivo)
	estaciónTemperatura := make(map[string][]float64)

	for {
		linea, err := reader.ReadBytes('\n')
		if err != nil {
			break
		}

		semicolonIndex := -1
		for i, char := range linea {
			if char == ';' {
				semicolonIndex = i
				break
			}
		}
		if semicolonIndex == -1 {
			break
		}

		nombre := string(linea[:semicolonIndex])
		numero, err := strconv.ParseFloat(string(linea[semicolonIndex+1:len(linea)-1]), 64)
		if err != nil {
			fmt.Println("Error al convertir la temperatura a número:", err)
			break
		}

		if array, exist := estaciónTemperatura[nombre]; !exist {
			estaciónTemperatura[nombre] = []float64{numero, numero, numero, 1.0}
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

	// Resto del código para ordenar y mostrar resultados
	var claves [][]byte
	for clave := range estaciónTemperatura {
		claves = append(claves, []byte(clave))
	}
	sort.Sort(ByteSlice(claves))

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

// ByteSlice es una implementación de la interfaz sort.Interface para slices de bytes.
type ByteSlice [][]byte

func (s ByteSlice) Len() int           { return len(s) }
func (s ByteSlice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s ByteSlice) Less(i, j int) bool { return string(s[i]) < string(s[j]) }

func Round(num float64) float64 {
	if num >= 0 {
		return math.Round(num*10.0) / 10.0
	}
	return math.RoundToEven(num*10.0) / 10.0
}
