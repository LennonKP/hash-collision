# v2: Busca Paralela (Multi-Thread Otimizada)

Esta versão foca em extrair o máximo de performance da CPU e otimizar o uso da memória RAM para permitir buscas em espaços maiores (até 56 bits).

## Otimizações Realizadas

### 1. Paralelismo com Goroutines
O programa utiliza `runtime.NumCPU()` para identificar o número de núcleos lógicos e lança Workers paralelos. Isso permite que a geração de hashes ocorra simultaneamente em todos os núcleos.

### 2. Mapas Particionados (Shards)
Para evitar que as threads "briguem" por um único cadeado (Mutex Contention) ao acessar o mapa, dividimos o armazenamento em **256 Shards**. Cada thread escreve em um shard diferente baseado no primeiro byte do hash, o que elimina gargalos de sincronização.

### 3. Sementes uint64 (Economia de 75% de RAM)
Em vez de salvar a `string` original (que é pesada), salvamos apenas uma "semente" numérica de 64 bits (`uint64`). 
- **Antes**: String (~60 bytes) + Overhead.
- **Agora**: uint64 (8 bytes) + Overhead.
Isso permite colocar quase 4x mais dados no mesmo pente de memória RAM.

## Estimativas de Recursos (CPU Multi-Core)

| Bits | Tentativas Médias | Memória (Otimizada) | Tempo (16 Cores) |
| :--- | :--- | :--- | :--- |
| 40 bits | ~ 1 milhão | ~ 30 MB | ~ 50 ms |
| 48 bits | ~ 16 milhões | ~ 400 MB | ~ 2 s |
| 56 bits | ~ 268 milhões | **~ 8-12 GB** | ~ 20-40 s |
| 64 bits | ~ 4.2 bilhões | **~ 180 GB+** | **Inviável (RAM)** |

> [!TIP]
> Esta versão mostra o limite do hardware doméstico. Conseguimos ocupar 100% da CPU, mas ainda esbarramos na barreira física da memória RAM em 64 bits.
