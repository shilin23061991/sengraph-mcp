# sentgraph-mcp

Go MCP-сервер долговременной памяти для кодинг-агентов на базе Zep Cloud.

Sentgraph держит локальный слой тонким:

- Zep Cloud отвечает за построение графа, дедупликацию, эмбеддинги и извлечение.
- Sentgraph отдаёт агенту шесть базовых MCP-инструментов.
- Нативные Go-хуки часто читают и пишут память без локального демона.
- Память проекта разделена по `project_id`, поэтому один проект может охватывать несколько репозиториев.

## Установка

```bash
go install github.com/shilin23061991/sentgraph-mcp@latest
```

## Установка в Claude Code

Шаги ниже подключают MCP-сервер, ставят скилы и хуки и дописывают промт.
 
Все команды выполняйте из корня репозитория.

### 1. Ключи окружения

`sentgraph-mcp` сам читает `.env.local` из корня проекта при старте (ищет вверх по дереву от `CLAUDE_PROJECT_DIR`/рабочего каталога), поэтому у каждого проекта свои ключи и общий глобальный env не нужен.

Создайте `.env.local` одной командой (он в `.gitignore`, в коммит не попадёт; значения замените на свои):

```bash
cat > .env.local <<'EOF'
ZEP_API_KEY=вашключ
ZEP_USER_ID=вашид
SENTGRAPH_PROJECT_ID=имя-проекта
EOF
```

Вручную экспортировать ничего не нужно. Приоритет стандартный (non-override): если переменная уже задана в окружении, она побеждает, а `.env.local` лишь заполняет недостающие.

### 2. Подключить MCP-сервер

stdio (по умолчанию для Claude Code). Скоуп проекта: конфиг пишется в `.mcp.json` в корне репозитория (коммитится, шарится с командой). Ключи передавать не нужно -- бинарь берёт их из `.env.local`:

```bash
claude mcp add --scope project --transport stdio sentgraph -- sentgraph-mcp serve
```

> В `.mcp.json` секретов нет -- все ключи бинарь читает из `.env.local` (шаг 1), поэтому файл безопасно коммитить. `--` отделяет имя сервера `sentgraph` от команды запуска `sentgraph-mcp serve`.
>
> Перед первым использованием project-сервер требует подтверждения (`claude mcp list` покажет `Pending approval`).

```bash
claude mcp list
claude mcp get sentgraph
```

### 3. Установить скилы

Личные (для всех проектов):

```bash
mkdir -p ~/.claude/skills
cp -R plugin/skills/* ~/.claude/skills/
```

Либо в рамках проекта (коммитятся в репозиторий):

```bash
mkdir -p .claude/skills
cp -R plugin/skills/* .claude/skills/
```

Станут доступны команды `/remember`, `/recall`, `/forget`, `/session-history`, `/sentgraph-tools`.

### 4. Установить хуки

Хуки объявляются в `settings.json`. Команда ниже вмердживает блок `hooks` из `plugin/hooks/hooks.json` в пользовательские настройки (нужен `jq`):

```bash
mkdir -p ~/.claude
[ -f ~/.claude/settings.json ] || echo '{}' > ~/.claude/settings.json
tmp="$(mktemp)"
jq --slurpfile h plugin/hooks/hooks.json '.hooks = ((.hooks // {}) * $h[0].hooks)' \
  ~/.claude/settings.json > "$tmp" && mv "$tmp" ~/.claude/settings.json
```

> Команда идемпотентна, но заменяет массивы хуков для событий `SessionStart`, `UserPromptSubmit`, `PreCompact`, `Stop`, `SessionEnd`. Если на этих событиях уже висят ваши хуки, объедините блоки вручную. Для настроек проекта используйте `.claude/settings.json`.

### 5. Дописать промт

Добавьте блок памяти в файл памяти Claude (один раз):

```bash
# либо для текущего проекта
printf '\n' >> ./CLAUDE.md && cat CLAUDE.snippet.md >> ./CLAUDE.md
```

### Финальная проверка

```bash
sentgraph-mcp doctor --online   # проверка конфигурации + связи с Zep
claude mcp list                 # sentgraph должен быть в списке
```

## Команды

```bash
sentgraph-mcp doctor             # проверка конфигурации (API-ключ, пользователь, project id)
sentgraph-mcp doctor --online    # дополнительно проверить связь с Zep (user/project graph/thread)
sentgraph-mcp serve              # MCP по stdio (по умолчанию для Claude Code / Cursor)
sentgraph-mcp serve --http :8080 # MCP по Streamable HTTP на ADDR
sentgraph-mcp hook SessionStart
```

## MCP-инструменты

- `memory_context` -- собранный контекст пользователя и проекта.
- `memory_search` -- поиск по графовой памяти пользователя или проекта.
- `memory_history` -- недавние сообщения треда.
- `memory_add_messages` -- сохранить ходы беседы.
- `memory_add` -- сохранить факты или данные проекта/пользователя.
- `memory_forget` -- удалить edge, node или episode по UUID.

## Хуки

Конфигурация хуков плагина Claude -- в `plugin/hooks/hooks.json`.

События по умолчанию:

- `SessionStart` -- прочитать контекст и подгрузить его.
- `UserPromptSubmit` -- записать промт пользователя и подгрузить свежий контекст.
- `PreCompact` -- повторно подгрузить контекст перед компакцией.
- `Stop` -- сохранить последний ход ассистента из транскрипта.
- `SessionEnd` -- финальный проход сохранения.
