package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
)

func main() {
	fmt.Println(createFinalString())
}

func createFinalString() string {

	tempMap, err := readFileLineByLineIntoAMap("measurements_50M.txt")
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

type minMaxSumCount struct {
	min   int64
	max   int64
	sum   int64
	count int64
}

func readFileLineByLineIntoAMap(path string) (map[string]minMaxSumCount, error) {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}

	chanOwner := func() <-chan []string {
		//We send a chunck of lines to reduce the number of items send on the channel
		//Buffered channel 100 lines
		resultChan := make(chan []string, 100)
		//Separating blocks of 100 lines
		toSend := make([]string, 100)

		//Making a go routine for reading lines and send it throw 'resultchan'
		go func() {
			defer close(resultChan)
			scanner := bufio.NewScanner(file)
			var count int
			//.Scan() reads the file until it finds a \n and returns it
			for scanner.Scan() {
				if count == 100 {
					//To avoid race condition we will need to create a copy of the slice to send it over to the channel
					//// Note: i think this happends becouse we are sending a referens of the slice into a channel ////
					localCopy := make([]string, 100)
					copy(localCopy, toSend)
					resultChan <- localCopy
					count = 0
				}
				//Adding lines to the sendArray
				toSend[count] = scanner.Text()
				count++
			}
			if count != 0 {
				resultChan <- toSend[:count]
			}
		}()
		return resultChan
	}

	resultChan := chanOwner()
	tempMap := make(map[string]minMaxSumCount)
	//For each array/slice in the channel
	for lineSlice := range resultChan {
		//For each line of the slice
		for _, line := range lineSlice {
			//Try to separete the city string to the temperature value
			index := strings.Index(line, ";")
			if index == -1 {
				continue
			}
			city := line[:index]
			temp := stringToInt64(line[index+1:])
			//Check if the city exist in the map
			//Is exist count++ sum+new value, and check if the new value is less than the min or greater than the max
			if num, ok := tempMap[city]; ok {
				num.count++
				num.sum += temp
				if temp < num.min {
					num.min = temp
				}
				if temp > num.max {
					num.max = temp
				}
				tempMap[city] = num
			} else {
				//Else create a new minMaxSumCount object
				tempMap[city] = minMaxSumCount{
					min:   temp,
					max:   temp,
					sum:   temp,
					count: 1,
				}
			}
		}
	}
	return tempMap, nil
}

func stringToInt64(input string) int64 {
	input = input[:len(input)-2] + input[len(input)-1:]
	output, _ := strconv.ParseInt(input, 10, 64)
	return output
}

func round(x float64) float64 {
	rounded := math.Round(x * 10)
	if rounded == -0.0 {
		return 0.0
	}
	return rounded / 10
}
