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
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	return memStats.Alloc
}

func PrintResults(result Result, title string) {
	bytesLen := result.Bits / 8
	fmt.Println("========================================")
	fmt.Printf("💥 Colisão Encontrada (%s)!\n", title)
	fmt.Printf("Mini-Hash (Hex): %0*x\n", bytesLen*2, result.MiniHash)
	fmt.Printf("String 1: %s\n", result.String1)
	fmt.Printf("String 2: %s\n", result.String2)
	fmt.Println("----------------------------------------")
	fmt.Printf("Tentativas totais: %d\n", result.Attempts)
	fmt.Printf("Tempo total de execução: %v\n", result.Duration)
	fmt.Printf("Memória RAM Inicial: %d MB\n", result.InitialMemory/1024/1024)
	fmt.Printf("Memória RAM Final: %d MB\n", result.FinalMemory/1024/1024)
	fmt.Println("========================================")
}

// CreateMask gera uma máscara de N bits (suporta até 64 bits)
func CreateMask(bits int) uint64 {
	if bits == 64 {
		return ^uint64(0)
	}
	return uint64((1 << bits) - 1)
}

// ExtractUint64 converte os primeiros N bytes de um array para um uint64
func ExtractUint64(h [32]byte, bytesLen int) uint64 {
	var val uint64
	for i := 0; i < bytesLen; i++ {
		val = (val << 8) | uint64(h[i])
	}
	return val
}

// IsDistinguished verifica se um hash atende ao critério de prefixo de zeros
func IsDistinguished(hash uint64, k int) bool {
	mask := CreateMask(k)
	return (hash & mask) == 0
}

// ApplyMask aplica uma máscara de bits a um valor uint64
func ApplyMask(val uint64, mask uint64) uint64 {
	return val & mask
}
