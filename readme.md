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
  - A tabela agora inclui a coluna "Data do Aporte".
    - Se "Aporte no início do mês" estiver marcado, a data mostrada é o primeiro dia do mês.
    - Se não estiver marcado, a data mostrada é o último dia do mês.
    - O mês 1 corresponde ao mês corrente.

Estrutura Web:
- CSS separado em `web/app.css`.
- JavaScript separado em `web/app.js` (inclui lógica de formulário, renderização, alternância de tema e exportação PDF/CSV).

Nota: a versão antiga baseada em TUI (tcell/tview) foi substituída pela interface web.