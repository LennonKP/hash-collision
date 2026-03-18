# Explicação Técnica: Otimizações e Algoritmos

Este documento detalha os conceitos avançados utilizados para atingir a marca de 64 bits no experimento de colisão de hash.

---

## 0. Por que usar um `map`? (A Busca por Colisão)

Para encontrar uma colisão, o programa precisa de **memória**. 
Ao gerar um novo Hash (ex: `H2`), como ele sabe que esse hash já apareceu antes? 
- **Sem o `map`**: Ele teria que comparar `H2` com cada um dos milhões de hashes gerados anteriormente (`H1`). Isso transformaria o algoritmo em algo extremamente lento ($O(n^2)$), pois a cada novo passo, a lista de comparação cresce.
- **Com o `map`**: O Go utiliza uma *Hash Table*. Isso permite que o programa pergunte: *"O `H2` já está aqui?"* e obtenha a resposta instantaneamente ($O(1)$), não importa se já temos 10 ou 10 milhões de hashes guardados.

O **`map`** é a nossa "memória fotográfica" que permite detectar o par repetido no momento exato em que ele surge.

---

## 1. O que são Shards e por que 256?

No Go, um `map` comum **não é seguro para uso concorrente**. Se duas threads tentarem escrever nele ao mesmo tempo, o programa sofre um *panic*. 

A solução inicial é usar um `sync.Mutex` (um cadeado):
- **Problema**: Se você tem 16 núcleos tentando escrever em um único mapa, 15 núcleos ficam parados esperando o cadeado abrir. Isso é chamado de **Mutex Contention**. Você tem 16 núcleos, mas a velocidade é limitada pela fila do cadeado.

**Por que 256?**
Dividimos o mapa em 256 "shards" (pedaços) independentes, cada um com seu próprio cadeado. Escolhemos 256 por três motivos:
1. **Poder de 2**: 256 é $2^8$. Computadores processam divisões e restos (módulo) por potências de 2 de forma quase instantânea.
2. **Probabilidade**: Com 16 threads e 256 opções, a chance de duas threads escolherem o mesmo caderno ao mesmo tempo é de apenas **0,4%**. Isso elimina virtualmente qualquer espera (contention).
3. **Equilíbrio**: É um número grande o suficiente para o seu Ultra 7, mas pequeno o suficiente para não gastar memória excessiva apenas criando os objetos de controle.

---

## 2. O Algoritmo de Pollard's Rho: Como ele detecta sem memória?

Diferente do dicionário (v1/v2), o Pollard's Rho só salva "Checkpoints" (Pontos Distinguidos).

### Passo a Passo da Detecção:
1. **A Caminhada**: Cada thread começa em um ponto aleatório (Semente A) e vai pulando de hash em hash: $Semente \rightarrow H1 \rightarrow H2 \rightarrow H3 \dots$
2. **O Checkpoint**: Ela só salva no mapa quando encontra um "Ponto Distinguido" (ex: um hash que começa com 16 zeros). Vamos supor que ela salvou o $H_{1000}$.
3. **O Encontro**: Outra thread (Semente B) está caminhando e, por coincidência, gera um hash que a Semente A já gerou antes. A partir desse momento, as duas trilhas se tornam **idênticas** (pois o Hash de um valor é sempre o mesmo).
4. **A Detecção**: A Semente B continuará caminhando até encontrar o seu próximo Ponto Distinguido. Como as trilhas se uniram, ela eventualmente chegará no **mesmo Ponto Distinguido** que a Semente A salvou ($H_{1000}$).
5. **O Alerta**: O programa vê: *"Opa! Esse ponto já foi atingido pela Semente A!"*. Isso é o sinal de que as trilhas colidiram em algum lugar do passado.
6. **O Backtrace (Volta no Tempo)**: O programa volta ao início (Semente A e Semente B) e caminha passo a passo comparando os resultados até encontrar o exato momento em que o Hash de A e o Hash de B se tornaram iguais. 

**Moral da história**: Nós não precisamos lembrar de todos os passos, apenas dos "Postos de Gasolina" (Pontos Distinguidos). Se dois carros param no mesmo posto vindo de cidades diferentes, sabemos que eles pegaram a mesma estrada em algum momento.

---

## 3. Como calcular as estimativas?

### A) Tentativas (Paradoxo do Aniversário)
A fórmula para o número esperado de tentativas ($E$) para encontrar uma colisão em um espaço de $N$ bits é:
$$E \approx \sqrt{\frac{\pi}{2} \cdot 2^N} \approx 1.25 \cdot 2^{N/2}$$

**Exemplo para 64 bits:**
- $N = 64 \rightarrow N/2 = 32$.
- $2^{32} = 4.294 .967 .296$.
- Estima-se $\approx 5.4$ bilhões de tentativas. (Exatamente o que vimos no seu log de 7 minutos!).

### B) Memória (RAM)
No Go, um mapa `map[uint64]uint64` gasta aproximadamente **24 a 32 bytes** por entrada (considerando overhead de buckets e ponteiros).

**Cálculo v2 (48 bits):**
- Tentativas: $\approx 16$ milhões.
- Memória: $16.000 .000 \times 24 \text{ bytes} \approx 384 \text{ MB}$. (Na prática sobe para ~1GB devido ao lixo e crescimento do mapa).

**Cálculo v1 (com Strings):**
- Uma string hexadecimal de 32 caracteres gasta 32 bytes + overhead do cabeçalho da string. O custo sobe para **~80 bytes** por entrada. Por isso a v1 trava muito antes.

### C) Tempo
O tempo é simplesmente: 
$$T = \frac{\text{Tentativas Totais}}{\text{Hashes por Segundo pelo Hardware}}$$

No seu **Ultra 7**, medimos que ele faz aproximadamente **12 a 15 milhões de hashes por segundo** (somando todos os núcleos).
- Para 64 bits (5.4 bilhões de tentativas):
- $5.400 .000 .000 / 13.000 .000 = \approx 415 \text{ segundos} \approx 6.9 \text{ minutos}$.

---

## 4. Como as strings aleatórias são geradas?

Em todas as versões, nós não usamos palavras do dicionário (como "senha123"), mas sim **dados puramente aleatórios**.

### O Motor de Caos: `crypto/rand`
Usamos o pacote `crypto/rand` do Go. Diferente de outros geradores, ele usa o hardware do seu notebook para criar números que são **impossíveis de prever**.

### Onde elas são criadas?
Elas são criadas **dentro do loop infinito** (`for { ... }`). A cada ciclo (iteração), o programa faz:

1.  **Criação**: O comando `rand.Read(buf)` "joga um dado" e preenche um espaço de memória com lixo aleatório.
2.  **Preparação**: 
    - Na **Versão 1**, transformamos esse lixo em uma `string` legível (ex: `"f8e2..."`).
    - Na **Versão 2**, transformamos em um número **`uint64`** (mais rápido e leve).
3.  **Comparação**: Calculamos o Hash desse valor novo e perguntamos ao mapa: *"Ei, alguém já passou por aqui e deixou um rastro igual a esse?"*. 
    - Se **Sim**: O mapa nos devolve o que estava guardado lá (`originalStr` ou `originalSeed`). Comparamos os dois e, se forem diferentes... **COLISÃO!** ✨
    - Se **Não**: Guardamos esse novo rastro no mapa e passamos para o próximo ciclo.

## 5. Dicionário de Operadores (O que é `<<` e `&`?)

Se você está começando no Go, esses símbolos parecem estranhos, mas eles são os segredos da alta performance:

### O Deslocamento (`<<` - Left Shift)
O símbolo `<<` empurra os bits para a esquerda. 
- **Exemplo**: `1 << 3`
- Em binário, o número `1` é `0001`.
- Empurrando 3 casas para a esquerda, ele vira `1000` (que é o número **8**).
- **Na prática**: é uma forma ultra-rápida de calcular potências de 2. Usamos `1 << bits` para saber o tamanho total do nosso espaço de busca.

### O E-Comercial (`&` - Bitwise AND)
O símbolo `&` funciona como uma **Peneira** ou **Máscara**:
- Ele compara dois números bit a bit. Só sobra `1` onde **ambos** forem `1`.
- **Exemplo**:
  - Hash original: `10110110`
  - Máscara (2 bits): `00000011`
  - Resultado: `00000010` (Sobrou apenas o que passou pelos "furos" da máscara).
- **Na prática**: Usamos isso para "cortar" o SHA-256 e ficar só com os bits que nos interessam para a colisão.

---

## 6. Por que extrair para `uint64`?

Você notou que no código fazemos `hashutils.ExtractUint64(fullHash, bytesLen)`. Isso é fundamental por três motivos técnicos:

1. **Velocidade de Comparação**:
   - O SHA-256 devolve um array de 32 bytes (`[32]byte`). Comparar dois arrays exige que o computador olhe cada byte um por um.
   - Comparar dois `uint64` é uma **instrução única** no seu processador Ultra 7. Como fazemos bilhões de comparações, isso economiza minutos de execução.

2. **Facilidade Matemática**:
   - No Go, você não consegue aplicar filtros como a Máscara (`&`) ou fazer deslocamentos (`<<`) diretamente em um array de bytes.
   - Ao transformar os bytes em um **Número**, ganhamos superpoderes matemáticos para "cortar" e "ajustar" o hash como quisermos.

3. **Eficiência no Mapa**:
   - O Go é extremamente otimizado para usar números simples (`uint64`) como chaves de busca em mapas. Usar arrays ou strings como chaves deixaria o programa muito mais pesado e lento.

**Resumo**: Nós transformamos os bytes em um número para "falar a língua nativa" do processador, garantindo que a busca seja a mais rápida possível.

---
