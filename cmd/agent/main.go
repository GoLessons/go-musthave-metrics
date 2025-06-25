package main

import (
	"fmt"
	"github.com/GoLessons/go-musthave-metrics/internal/agent/reader"
	"github.com/GoLessons/go-musthave-metrics/internal/server/storage"
	"github.com/GoLessons/go-musthave-metrics/pkg"
)

func main() {
	//var memStats runtime.MemStats
	var gaugeStorage pkg.Storage[float64]
	gaugeStorage = storage.NewMemStorage[float64]()

	//runtime.ReadMemStats(&memStats)

	memStatReader := reader.NewMemStatsReader()
	memStatReader.Refresh()

	val, ok := memStatReader.Get("Alloc")
	if !ok {
		fmt.Println("Cannot read metric: " + "Alloc")
	}

	err := gaugeStorage.Set("Alloc", val)
	if err != nil {
		fmt.Println(err)
	}

	// Получаем информацию об аллокациях
	fmt.Printf("Общее выделенное пространство памяти (байты): %f\n", val)
	/*fmt.Printf("Память в хипе (байты): %d\n", memStats.HeapAlloc)
	fmt.Printf("Количество сборок мусора: %d\n", memStats.NumGC)
	fmt.Printf("Высвобожденная память (байты): %d\n", memStats.HeapReleased)*/
}
