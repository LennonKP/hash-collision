package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"github.com/lennonkp/hash-collision/pkg/hashutils"
	"time"
)

func main() {
	var bits int
	flag.IntVar(&bits, "bits", 24, "Tamanho do Mini-Hash em bits (deve ser múltiplo de 8, máx 64)")
	flag.Parse()

	if err := hashutils.ValidateInput(bits); err != nil {
		fmt.Printf("Erro: %v\n", err)
		return
	}

	fmt.Printf("Iniciando busca por colisão (Single-Thread) para %d bits...\n", bits)

	result := findCollision(bits)
	hashutils.PrintResults(result, "Single-Thread")
}

func findCollision(bits int) hashutils.Result {
	bytesLen := bits / 8
	mask := hashutils.CreateMask(bits)

	hashes := make(map[uint64]string)

	initialMem := hashutils.GetAllocatedMemory()
	startTime := time.Now()
	var attempts int64 = 0
	buf := make([]byte, 16)

	for {
		attempts++

		rand.Read(buf)
		inputStr := hex.EncodeToString(buf)

		fullHash := sha256.Sum256([]byte(inputStr))
		hashUint := hashutils.ExtractUint64(fullHash, bytesLen)
		miniHash := hashutils.ApplyMask(hashUint, mask)

		if originalStr, exists := hashes[miniHash]; exists {
			if originalStr != inputStr {
				return hashutils.Result{
					Bits:          bits,
					Attempts:      attempts,
					Duration:      time.Since(startTime),
					InitialMemory: initialMem,
					FinalMemory:   hashutils.GetAllocatedMemory(),
					String1:       originalStr,
					String2:       inputStr,
					MiniHash:      miniHash,
				}
			}
		}
		hashes[miniHash] = inputStr
	}
}
