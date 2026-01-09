# Calculadora de investimentos CDI (versão Web)
## O uso é pessoal, mas se te for útil fica à vontade para usar

Agora a aplicação roda como um servidor HTTP e a interface é HTML.

Como executar:

1. Instale o Go (1.20+ recomendado).
2. No diretório do projeto, rode:
   
   ```bash
   go run .
   ```
3. Abra no navegador: http://localhost:8080

Recursos:
- Endpoint POST `/simulate` recebe os parâmetros em JSON e retorna o cálculo completo.
- Interface web em `web/index.html` com formulário e tabela de evolução.

Nota: a versão antiga baseada em TUI (tcell/tview) foi substituída pela interface web.