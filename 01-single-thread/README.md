# v1: Busca por Colisão (Single-Thread)

Esta é a versão inicial e mais simples do projeto, servindo como "Prova de Conceito" (PoC) para demonstrar o **Paradoxo do Aniversário**.

## Funcionamento Técnico
O algoritmo gera strings aleatórias e calcula o seu hash truncado de `N` bits. Cada hash é armazenado em um mapa (`map[uint64]string`) como chave, e a string original como valor.
- Antes de inserir, o programa verifica se a chave já existe.
- Se existir e as strings forem diferentes, uma **colisão** foi encontrada.

## Limitações e Gargalos
1.  **Thread Única**: Utiliza apenas um núcleo da CPU, ignorando o poder de processamento paralelo dos computadores modernos.
2.  **Consumo de RAM**: Como armazena a `string` hexadecimal completa para cada tentativa, o consumo de memória escala muito rápido.
3.  **Performance do Mapa**: Em Golang, mapas grandes em thread única começam a sofrer latência de inserção à medida que crescem.

## Estimativas de Recursos

| Bits | Tentativas Médias ($\sqrt{2^N}$) | Memória Estimada | Tempo Estimado (1 Thread) |
| :--- | :--- | :--- | :--- |
| 24 bits | 4.096 | < 1 MB | < 2 ms |
| 32 bits | 65.536 | ~ 5 MB | ~ 20 ms |
| 40 bits | 1.048.576 | ~ 120 MB | ~ 500 ms |
| 48 bits | 16.777.216 | **~ 2.5 GB** | ~ 8-15 s |
| 56 bits | 268.435.456 | **~ 40 GB** | **Inviável (RAM)** |

> [!IMPORTANT]
> Esta versão é excelente para entender o conceito, mas falha em bits altos devido ao armazenamento ineficiente de strings na memória RAM.
