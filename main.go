package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	// Задаем колличество воркеров
	var numWorkers int
	fmt.Print("Введите количество воркеров: ")
	fmt.Scan(&numWorkers)

	// Канал
	workerChannel := make(chan int)

	// Контекст для управления завершением
	ctx, cancel := context.WithCancel(context.Background())

	// Группа ожидания для воркеров
	var workerGroup sync.WaitGroup

	// Запуск воркеров
	for i := 1; i <= numWorkers; i++ {
		workerGroup.Add(1)
		go worker(ctx, i, workerChannel, &workerGroup)
	}

	// Генерация данных в главном потоке
	go func() {
		for {
			select {
			case <-ctx.Done():
				close(workerChannel)
				return
			default:
				workerChannel <- rand.Intn(100)
				time.Sleep(500 * time.Millisecond)
			}
		}
	}()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	//Завершение через ctrl+c
	<-signalChan
	fmt.Println("\nПолучен сигнал завершения. Останавливаем воркеров...")
	cancel()
	workerGroup.Wait()
	fmt.Println("Все воркеры завершили работу.")
}

func worker(ctx context.Context, id int, workerChannel <-chan int, workerGroup *sync.WaitGroup) {
	defer workerGroup.Done()
	for {
		select {
		case <-ctx.Done(): // Проверяем, если контекст завершен, то выходим
			fmt.Printf("Воркер %d завершает работу.\n", id)
			return
		case data, ok := <-workerChannel:
			if !ok {
				return
			}
			fmt.Printf("Воркер %d обработал: %d\n", id, data)
		}
	}
}
