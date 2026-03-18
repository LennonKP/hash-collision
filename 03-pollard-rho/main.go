package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"flag"
	"fmt"
	"github.com/lennonkp/hash-collision/pkg/hashutils"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

type DistPoint struct {
	StartSeed uint64
	Steps     int
}

func main() {
	var bits int
	var workers int
	var k int
	flag.IntVar(&bits, "bits", 24, "Tamanho do Mini-Hash em bits (múltiplo de 8)")
	flag.IntVar(&workers, "workers", runtime.NumCPU(), "Número de threads paralelas")
	flag.IntVar(&k, "k", 16, "Dificuldade do Ponto Distinguido (bits em zero no início)")
	flag.Parse()

	fmt.Printf("Iniciando Parallel Collision Search (PCS) para %d bits...\n", bits)
	fmt.Printf("Configuração: %d workers, Pontos Distinguidos: primeiros %d bits em zero\n", workers, k)
	fmt.Println("Vantagem: Memória quase zero (apenas pontos distinguidos são salvos)")

	startTime := time.Now()
	var totalAttempts int64

	distMap := make(map[uint64]DistPoint)
	var mapMutex sync.Mutex

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bytesLen := bits / 8
	mask := hashutils.CreateMask(bits)


	for i := 0; i < workers; i++ {
		go func() {
			var seedBuf [8]byte
			for {
				select {
				case <-ctx.Done():
					return
				default:
					// 1. Iniciar nova trilha com semente aleatória
					rand.Read(seedBuf[:])
					startSeed := binary.LittleEndian.Uint64(seedBuf[:])
					currentSeed := startSeed
					
					steps := 0
					for {
						steps++
						atomic.AddInt64(&totalAttempts, 1)

						// 2. Função iterativa: x_{i+1} = Hash(x_i)
						var currentBuf [8]byte
						binary.LittleEndian.PutUint64(currentBuf[:], currentSeed)
						fullHash := sha256.Sum256(currentBuf[:])
						
						hashUint := hashutils.ExtractUint64(fullHash, bytesLen)
						miniHash := hashutils.ApplyMask(hashUint, mask)

						// 3. Verificar se caiu num ponto distinguido
						if hashutils.IsDistinguished(miniHash, k) {
							mapMutex.Lock()
							if prev, exists := distMap[miniHash]; exists {
								if prev.StartSeed != startSeed {
									// POSSÍVEL COLISÃO ENCONTRADA!
									// Precisamos confirmar se nao foi uma "colisao de ponto" ou "colisao real"
									s1, s2, ok := findCollisionInChains(prev.StartSeed, startSeed, prev.Steps, steps, bytesLen, mask)
									if ok {
										duration := time.Since(startTime)
										fmt.Println("\n========================================")
										fmt.Println("💥 COLISÃO ENCONTRADA (Pollard's Rho / PCS)!")
										fmt.Printf("Trilha 1 (Início): %016x\n", prev.StartSeed)
										fmt.Printf("Trilha 2 (Início): %016x\n", startSeed)
										fmt.Println("----------------------------------------")
										fmt.Printf("String 1: %s\n", s1)
										fmt.Printf("String 2: %s\n", s2)
										fmt.Printf("Ambas resultam no Hash: %s\n", computeHashStr(s1, bytesLen, mask))
										fmt.Println("----------------------------------------")
										fmt.Printf("Tentativas totais: %d\n", atomic.LoadInt64(&totalAttempts))
										fmt.Printf("Tempo: %v\n", duration)
										fmt.Printf("Pontos Distinguidos Salvos: %d\n", len(distMap))
										fmt.Println("========================================")
										cancel()
										mapMutex.Unlock()
										return
									}
								}
							} else {
								distMap[miniHash] = DistPoint{StartSeed: startSeed, Steps: steps}
							}
							mapMutex.Unlock()
							break // Inicia nova trilha após ponto distinguido
						}
						
						currentSeed = miniHash // Próximo passo da corrente
						
						// Proteção contra trilhas infinitas (raro mas possível)
						if steps > 10000000 {
							break
						}
					}
				}
			}
		}()
	}

	<-ctx.Done()
}

// findCollisionInChains reconstrói as duas trilhas para encontrar o ponto exato onde elas colidiram
func findCollisionInChains(seed1, seed2 uint64, steps1, steps2 int, bytesLen int, mask uint64) (string, string, bool) {
	// 1. Alinhar as trilhas (avançar a mais longa até ficarem com mesmo tamanho restante)
	curr1 := seed1
	curr2 := seed2
	
	if steps1 > steps2 {
		for i := 0; i < steps1-steps2; i++ {
			curr1 = step(curr1, bytesLen, mask)
		}
	} else {
		for i := 0; i < steps2-steps1; i++ {
			curr2 = step(curr2, bytesLen, mask)
		}
	}

	// 2. Avançar juntas até encontrar o ponto de encontro
	var prev1, prev2 uint64
	for {
		if curr1 == curr2 {
			// A colisão ocorreu um passo ANTES do ponto de encontro
			if prev1 != prev2 && prev1 != 0 {
				return fmt.Sprintf("%016x", prev1), fmt.Sprintf("%016x", prev2), true
			}
			return "", "", false // "Robin Hood" ou colisão no início
		}
		prev1, prev2 = curr1, curr2
		curr1 = step(curr1, bytesLen, mask)
		curr2 = step(curr2, bytesLen, mask)
	}
}

func step(seed uint64, bytesLen int, mask uint64) uint64 {
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], seed)
	fullHash := sha256.Sum256(buf[:])
	hashUint := hashutils.ExtractUint64(fullHash, bytesLen)
	return hashutils.ApplyMask(hashUint, mask)
}

func computeHashStr(seedHex string, bytesLen int, mask uint64) string {
	decodedSeed, _ := hex.DecodeString(seedHex)
	fullHash := sha256.Sum256(decodedSeed)
	hashUint := hashutils.ExtractUint64(fullHash, bytesLen)
	return fmt.Sprintf("%0*x", bytesLen*2, hashutils.ApplyMask(hashUint, mask))
}
