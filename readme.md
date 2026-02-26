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

## Salvando simulações em Supabase

É possível gravar cada requisição/resultado na sua instância do Supabase. Crie a tabela SQL
(rodando em `SQL editor` do Supabase ou via `psql`):

```sql
create table simulations (
  id uuid primary key default gen_random_uuid(),
  user_id uuid references auth.users(id),
  request jsonb,
  response jsonb,
  created_at timestamptz default now()
);
```

Depois configure as variáveis de ambiente no mesmo ambiente onde o servidor Go roda:

```bash
export SUPABASE_URL="https://your-project.supabase.co"
export SUPABASE_KEY="<anon-or-service-role-key>"
# opcional: id do usuário atual, caso utilize autenticação
export SUPABASE_USER_ID="00000000-0000-0000-0000-000000000000"
```

A aplicação detecta as variáveis (incluindo via arquivo `.env` se existir na raiz, pois o
módulo `joho/godotenv` é usado durante a inicialização) e, após cada simulação bem‑sucedida,
envia um `POST` para `/rest/v1/simulations`. Erros de gravação são apenas logados localmente, sem
interferir na API pública.

Um exemplo de `.env`:

```env
SUPABASE_URL="https://your-project.supabase.co"
SUPABASE_KEY="<your-key>"
SUPABASE_USER_ID="00000000-0000-0000-0000-000000000000"
```

```bash
# rodando com as variáveis configuradas:
go run .
```

(Dica: use uma [chave de serviço](https://supabase.com/docs/guides/api#service-role) se
quiser permitir gravações sem expor o valor em browsers.)

Estrutura Web:
- CSS separado em `web/app.css`.
- JavaScript separado em `web/app.js` (inclui lógica de formulário, renderização, alternância de tema e exportação PDF/CSV).

Nota: a versão antiga baseada em TUI (tcell/tview) foi substituída pela interface web.