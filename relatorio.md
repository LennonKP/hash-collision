# Atividade: Experimento de Colisão de Mini-Hash (Resultado Final - Alta Performance)

Este documento apresenta os resultados finais da atividade de experimentação algorítmica. O projeto evoluiu de uma busca simples em memória para um sistema de alta performance capaz de quebrar colisões de 64 bits em hardware doméstico utilizando algoritmos avançados.

## Resultados da Coleta de Dados

Implementamos duas abordagens:
1.  **Dicionário Otimizado (Sharded Maps)**: Rápido, mas limitado pela RAM.
2.  **Parallel Collision Search (Pollard's Rho / PCS)**: Extremamente eficiente em memória, permitindo atingir 64 bits.

Abaixo está a tabela consolidada de resultados finais obtidos no hardware **Intel Ultra 7 (16 threads)**:

| Tamanho (Bits) | Algoritmo | Tentativas totais | Tempo | RAM |
| :--- | :--- | :--- | :--- | :--- |
| 24 bits | Map Simples | ~ 8.000 | < 1 ms | ~ 1 MB |
| 32 bits | Map Simples | ~ 34.000 | ~ 13 ms | 3 MB |
| 40 bits | Map Simples | ~ 2.000.000 | ~ 200 ms | 200 MB |
| 48 bits | Sharded Map | 57.025.138 | 40.9 s | 4.1 GB |
| 56 bits | Sharded Map | 231.237.108 | 26.1 s | 6.9 GB |
| **64 bits** | **Pollard's Rho** | **5.411.041.349** | **7m 15s** | **< 1 MB** (1.232 pts) |

*Nota: A execução de 64 bits com **k=22** salvou apenas **1.232 pontos distinguidos**, ocupando uma fração minúscula de memória, enquanto processou mais de 5,4 bilhões de hashes.*

## Análise e Reflexão

### 1. Escalabilidade: Linear vs Exponencial
O experimento demonstrou de forma prática o crescimento **exponencial** da dificuldade. 
- De **56 bits** para **64 bits**, o espaço de busca aumentou 256 vezes.
- O tempo de processamento saltou de alguns segundos para **6 minutos**.
- A abordagem inicial de dicionário tornou-se inviável (o computador travava), provando que o **espaço de memória** escala de forma ainda mais crítica que o poder de processamento em computadores comuns.

### 2. Reflexão sobre Segurança: Por que o MD5 (128 bits) é vulnerável?
Mesmo que **128 bits** ($2^{128}$) pareça um número inalcançável, o **Paradoxo do Aniversário** reduz a segurança real para o patamar de $2^{64}$. 

Como demonstrado pela nossa execução final, um processador de notebook moderno (Ultra 7) conseguiu atravessar o espaço de $2^{32}$ (o "alvo" para colisão de 64-bits) em apenas 6 minutos. Um atacante com uma rede de GPUs ou hardware especializado que processe trilhões de hashes por segundo conseguiria atingir a marca de $2^{64}$ (colisão do MD5) em um tempo factível. Isso torna o MD5 totalmente obsoleto para garantir integridade diante de adversários com poder de processamento paralelo.
