package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"runtime"
	"time"
)

type Result struct {
	Bits          int
	Attempts      int
	Duration      time.Duration
	InitialMemory uint64
	FinalMemory   uint64
	String1       string
	String2       string
	MiniHash      uint64
}

func main() {
	var bits int
	flag.IntVar(&bits, "bits", 24, "Tamanho do Mini-Hash em bits (deve ser múltiplo de 8, máx 64)")
	flag.Parse()

	if err := validateInput(bits); err != nil {
		fmt.Printf("Erro: %v\n", err)
		return
	}

	fmt.Printf("Iniciando busca por colisão para %d bits...\n", bits)

	result := findCollision(bits)
	printResults(result)
}

func validateInput(bits int) error {
	if bits%8 != 0 || bits <= 0 || bits > 64 {
		return fmt.Errorf("o tamanho em bits deve ser > 0, <= 64 e múltiplo de 8")
	}
	return nil
}

func findCollision(bits int) Result {
	bytesLen := bits / 8
	mask := uint64((1 << bits) - 1)
	if bits == 64 {
		mask = ^uint64(0)
	}

	hashes := make(map[uint64]string)

	initialMem := getAllocatedMemory()
	startTime := time.Now()
	attempts := 0
	buf := make([]byte, 16)

	for {
		attempts++
		
		rand.Read(buf)
		inputStr := hex.EncodeToString(buf)

		h := sha256.Sum256([]byte(inputStr))
		var hashUint uint64
		for i := 0; i < bytesLen; i++ {
			hashUint = (hashUint << 8) | uint64(h[i])
		}
		miniHash := hashUint & mask

		if originalStr, exists := hashes[miniHash]; exists {
			if originalStr != inputStr {
				return Result{
					Bits:          bits,
					Attempts:      attempts,
					Duration:      time.Since(startTime),
					InitialMemory: initialMem,
					FinalMemory:   getAllocatedMemory(),
					String1:       originalStr,
					String2:       inputStr,
					MiniHash:      miniHash,
				}
			}
		}

		hashes[miniHash] = inputStr
	}
}

func getAllocatedMemory() uint64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m.Alloc
}

func printResults(r Result) {
	bytesLen := r.Bits / 8
	fmt.Println("========================================")
	fmt.Println("💥 Colisão Encontrada!")
	fmt.Printf("Mini-Hash (Hex): %0*x\n", bytesLen*2, r.MiniHash)
	fmt.Printf("String 1: %s\n", r.String1)
	fmt.Printf("String 2: %s\n", r.String2)
	fmt.Println("----------------------------------------")
	fmt.Printf("Tentativas até a colisão: %d\n", r.Attempts)
	fmt.Printf("Tempo total de execução: %v\n", r.Duration)
	fmt.Printf("Memória RAM Inicial: %d MB\n", r.InitialMemory/1024/1024)
	fmt.Printf("Memória RAM Final: %d MB\n", r.FinalMemory/1024/1024)
	fmt.Println("========================================")
}
