package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
)

func main() {
	fmt.Printf(evaluate("measurements_50M.txt"))
}

func evaluate(file string) string {
	tempMap, err := readFileLineByLineIntoAMap(file)
	if err != nil {
		panic(err)
	}

	resultArr := make([]string, len(tempMap))
	var count int
	for city, _ := range tempMap {
		resultArr[count] = city
		count++
	}

	sort.Strings(resultArr)

	var stringsBuilder strings.Builder
	for _, i := range resultArr {
		stringsBuilder.WriteString(fmt.Sprintf("%s=%.1f/%.1f/%.1f, ", i,
			round(float64(tempMap[i].min)/10.0),
			round(float64(tempMap[i].sum)/10.0/float64(tempMap[i].count)),
			round(float64(tempMap[i].max)/10.0)))
	}
	return stringsBuilder.String()[:stringsBuilder.Len()-2]
}

type cityTemperatureInfo struct {
	count int64
	min   int64
	max   int64
	sum   int64
}

func readFileLineByLineIntoAMap(filepath string) (map[string]cityTemperatureInfo, error) {
	file, err := os.Open(filepath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	tempMap := make(map[string]cityTemperatureInfo)
	resultChan := make(chan map[string]cityTemperatureInfo, 10)
	chunkChan := make(chan []byte, 15)
	chunkSize := 64 * 1024 * 1024
	var wg sync.WaitGroup

	// spawn workers to consume (process) file chunks read
	for i := 0; i < runtime.NumCPU()-1; i++ {
		wg.Add(1)
		go func() {
			for chunk := range chunkChan {
				processReadChunk(chunk, resultChan)
			}
			wg.Done()
		}()
	}

	// spawn a goroutine to read file in chunks and send it to the chunk channel for further processing
	go func() {
		buf := make([]byte, chunkSize)
		remains := make([]byte, 0, chunkSize)
		for {
			readChunk, err := file.Read(buf)
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				panic(err)
			}
			buf = buf[:readChunk]

			toSend := make([]byte, readChunk)
			copy(toSend, buf)

			lastNewLineIndex := bytes.LastIndex(buf, []byte{'\n'})

			toSend = append(remains, buf[:lastNewLineIndex+1]...)
			remains = make([]byte, len(buf[lastNewLineIndex+1:]))
			copy(remains, buf[lastNewLineIndex+1:])

			chunkChan <- toSend

		}
		close(chunkChan)

		// wait for all chunks to be proccessed before closing the result stream
		wg.Wait()
		close(resultChan)
	}()

	// process all city temperatures derived after processing the file chunks
	for t := range resultChan {
		for city, tempInfo := range t {
			if val, ok := tempMap[city]; ok {
				val.count += tempInfo.count
				val.sum += tempInfo.sum
				if tempInfo.min < val.min {
					val.min = tempInfo.min
				}

				if tempInfo.max > val.max {
					val.max = tempInfo.max
				}
				tempMap[city] = val
			} else {
				tempMap[city] = tempInfo
			}
		}
	}

	return tempMap, nil
}

func convertStringToInt64(input string) int64 {
	input = input[:len(input)-2] + input[len(input)-1:]
	output, _ := strconv.ParseInt(input, 10, 64)
	return output
}

func processReadChunk(buf []byte, resultChan chan<- map[string]cityTemperatureInfo) {
	var stringsBuilder strings.Builder
	toSend := make(map[string]cityTemperatureInfo)
	var city string

	for _, char := range buf {
		if char == ';' {
			city = stringsBuilder.String()
			stringsBuilder.Reset()
		} else if char == '\n' {
			if stringsBuilder.Len() != 0 && len(city) != 0 {
				temp := convertStringToInt64(stringsBuilder.String())
				stringsBuilder.Reset()

				if val, ok := toSend[city]; ok {
					val.count++
					val.sum += temp
					if temp < val.min {
						val.min = temp
					}

					if temp > val.max {
						val.max = temp
					}
					toSend[city] = val
				} else {
					toSend[city] = cityTemperatureInfo{
						count: 1,
						min:   temp,
						max:   temp,
						sum:   temp,
					}
				}
			}
		} else {
			stringsBuilder.WriteByte(char)
		}
	}
	resultChan <- toSend
}

func round(x float64) float64 {
	rounded := math.Round(x * 10)
	if rounded == -0.0 {
		return 0.0
	}
	return rounded / 10
}
