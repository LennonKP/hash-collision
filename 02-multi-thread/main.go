package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"flag"
	"fmt"
	"github.com/lennonkp/hash-collision/pkg/hashutils"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

const numShards = 256

type shard struct {
	sync.Mutex
	m map[uint64]uint64 // Chave: Hash, Valor: Semente (Randon Payload)
}

func main() {
	var bits int
	var workers int
	flag.IntVar(&bits, "bits", 24, "Tamanho do Mini-Hash em bits (deve ser múltiplo de 8, máx 64)")
	flag.IntVar(&workers, "workers", runtime.NumCPU(), "Número de threads (goroutines) paralelas")
	flag.Parse()

	if err := hashutils.ValidateInput(bits); err != nil {
		fmt.Printf("Erro: %v\n", err)
		return
	}

	fmt.Printf("Iniciando busca ultra-otimizada (%d workers, %d shards) por colisão para %d bits...\n", workers, numShards, bits)

	result := findCollisionSharded(bits, workers)
	hashutils.PrintResults(result, "Ultra-Otimizada via uint64")
}

// findCollisionSharded utiliza sementes uint64 para economizar ~70%% de RAM em cada entrada
func findCollisionSharded(bits int, numWorkers int) hashutils.Result {
	bytesLen := bits / 8
	mask := hashutils.CreateMask(bits)

	shards := make([]*shard, numShards)
	for i := 0; i < numShards; i++ {
		shards[i] = &shard{m: make(map[uint64]uint64)}
	}
	
	initialMem := hashutils.GetAllocatedMemory()
	startTime := time.Now()
	var totalAttempts int64
	
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	resultChan := make(chan hashutils.Result, 1)
	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var seedBuf [8]byte
			for {
				select {
				case <-ctx.Done():
					return
				default:
					atomic.AddInt64(&totalAttempts, 1)
					
					// 1. Gerar semente aleatória (8 bytes)
					rand.Read(seedBuf[:])
					currentSeed := binary.LittleEndian.Uint64(seedBuf[:])

					// 2. Calcular SHA-256 da semente
					fullHash := sha256.Sum256(seedBuf[:])
					
					hashUint := hashutils.ExtractUint64(fullHash, bytesLen)
					miniHash := hashutils.ApplyMask(hashUint, mask)

					// 3. Verificar colisão no shard correto
					shardIdx := int(fullHash[0]) % numShards
					selectedShard := shards[shardIdx]

					selectedShard.Lock()
					if originalSeed, exists := selectedShard.m[miniHash]; exists {
						if originalSeed != currentSeed {
							cancel() 

							// Prepara as strings originais para exibicao
							var seed1Buf [8]byte
							binary.LittleEndian.PutUint64(seed1Buf[:], originalSeed)
							
							resultChan <- hashutils.Result{
								Bits:          bits,
								Attempts:      atomic.LoadInt64(&totalAttempts),
								Duration:      time.Since(startTime),
								InitialMemory: initialMem,
								FinalMemory:   hashutils.GetAllocatedMemory(),
								String1:       fmt.Sprintf("%x", seed1Buf),
								String2:       fmt.Sprintf("%x", seedBuf),
								MiniHash:      miniHash,
							}
							selectedShard.Unlock()
							return
						}
					} else {
						selectedShard.m[miniHash] = currentSeed
					}
					selectedShard.Unlock()
				}
			}
		}()
	}

	res := <-resultChan
	wg.Wait()
	return res
}
