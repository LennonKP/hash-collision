package hashutils

import (
	"fmt"
	"runtime"
	"time"
)

type Result struct {
	Bits          int
	Attempts      int64
	Duration      time.Duration
	InitialMemory uint64
	FinalMemory   uint64
	String1       string
	String2       string
	MiniHash      uint64
}

func ValidateInput(bits int) error {
	if bits%8 != 0 || bits <= 0 || bits > 64 {
		return fmt.Errorf("o tamanho em bits deve ser > 0, <= 64 e múltiplo de 8")
	}
	return nil
}

func GetAllocatedMemory() uint64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m.Alloc
}

func PrintResults(r Result, title string) {
	bytesLen := r.Bits / 8
	fmt.Println("========================================")
	fmt.Printf("💥 Colisão Encontrada (%s)!\n", title)
	fmt.Printf("Mini-Hash (Hex): %0*x\n", bytesLen*2, r.MiniHash)
	fmt.Printf("String 1: %s\n", r.String1)
	fmt.Printf("String 2: %s\n", r.String2)
	fmt.Println("----------------------------------------")
	fmt.Printf("Tentativas totais: %d\n", r.Attempts)
	fmt.Printf("Tempo total de execução: %v\n", r.Duration)
	fmt.Printf("Memória RAM Inicial: %d MB\n", r.InitialMemory/1024/1024)
	fmt.Printf("Memória RAM Final: %d MB\n", r.FinalMemory/1024/1024)
	fmt.Println("========================================")
}
