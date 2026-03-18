# v3: PCS (Parallel Collision Search) - Pollard's Rho

A versão definitiva do projeto. Utiliza o algoritmo de busca paralela de colisões de van Oorschot e Wiener, permitindo quebrar hashes de 64 bits com **memória quase zero**.

## O Pulo do Gato: Pontos Distinguidos

Em vez de salvar todos os hashes (bilhões de registros), esta versão utiliza a técnica de **Pontos Distinguidos**:
1.  As threads geram "correntes" onde um hash serve de semente para o próximo.
2.  O programa define um critério de raridade (ex: os primeiros `k` bits do hash devem ser zero).
3.  **Somente os pontos que atendem ao critério são salvos.**
4.  Se duas trilhas diferentes colidirem em qualquer ponto, elas eventualmente chegarão ao **mesmo** ponto distinguido.
5.  O algoritmo detecta o encontro e faz o "backtrack" para achar o par original da colisão.

## Otimizações de Hardware
- **Zero Swap**: Como usa pouquíssima RAM, o notebook não trava com o Swap do disco.
- **CPU Bound**: A performance depende 100% da velocidade da CPU (SHA-NI instructions).

## Estimativas e Resultados Reais

| Bits | Tentativas Reais | Pontos Salvos | Memória | Tempo (Ultra 7) |
| :--- | :--- | :--- | :--- | :--- |
| 56 bits | ~ 230 milhões | ~ 3.500 | < 5 MB | ~ 20 s |
| 64 bits | **~ 4.5 bilhões** | **~ 1.200** | **~ 10 MB** | **~ 6-7 min** |

### O Fator 'k'
O parâmetro `-k` ajusta o equilíbrio:
- **`k` alto**: Menos RAM usada, mas trilhas mais longas (mais tempo para detectar a colisão).
- **`k` baixo**: Mais RAM usada, detecção de colisão mais rápida após o cruzamento das trilhas.

> [!IMPORTANT]
> Esta versão prova que, com o algoritmo certo, um notebook comum pode quebrar e encontrar colisões em espaços de 64 bits que seriam impossíveis via dicionário simples.
