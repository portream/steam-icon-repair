# Steam Icon Repair

[Русский](#русский) · [English](#english)

---

## Русский

Маленькая консольная утилита для Windows, которая восстанавливает пропавшие `.ico`-иконки ярлыков Steam-игр на рабочем столе.

### Зачем это нужно

Steam при создании ярлыка игры на рабочем столе сохраняет `.ico`-файл в `<SteamPath>\steam\games\<hash>.ico`. Этот файл может пропасть по разным причинам:

- антивирус удалил «подозрительный» бинарный файл из системной папки;
- очистка диска или сторонняя «оптимизаторская» утилита снесла кэш;
- баг при обновлении Steam;
- ручная чистка папки `steam\games` пользователем.

Результат: на рабочем столе остаётся ярлык игры с дефолтной системной иконкой. `Steam Icon Repair` находит такие ярлыки и докачивает оригинальные иконки с публичного CDN Valve.

### Как это работает

Каждый ярлык Steam-игры — это `.url`-файл с содержимым вида:

```ini
[InternetShortcut]
URL=steam://rungameid/440
IconFile=C:\Program Files (x86)\Steam\steam\games\b32f4...c8.ico
IconIndex=0
```

Из этого файла извлекаются два значения:

- **`appid`** — из строки `URL=steam://rungameid/<appid>`;
- **`hash`** — имя `.ico`-файла без расширения из строки `IconFile=`.

Имея пару `(appid, hash)`, утилита строит публичный URL и качает иконку напрямую с CDN Steam:

```
https://cdn.cloudflare.steamstatic.com/steamcommunity/public/images/apps/<appid>/<hash>.ico
```

Файл сохраняется по исходному пути из `IconFile=`. Никаких ключей API не требуется — это публичные ассеты Steam Community.

### Где утилита ищет ярлыки

- `%USERPROFILE%\Desktop`
- `%PUBLIC%\Desktop`
- `%APPDATA%\Microsoft\Windows\Start Menu\Programs`
- `%PROGRAMDATA%\Microsoft\Windows\Start Menu\Programs`

Рекурсивно, по расширению `.url`.

### Установка

#### Вариант 1: готовый бинарник

Скачайте `steam-icon-repair.exe` из раздела [Releases](../../releases) и положите в любую папку.

#### Вариант 2: сборка из исходников

Требуется Go 1.21+.

```bash
git clone https://github.com/<user>/steam-icon-repair.git
cd steam-icon-repair
go build -o steam-icon-repair.exe
```

Кросс-компиляция из macOS/Linux:

```bash
GOOS=windows GOARCH=amd64 go build -o steam-icon-repair.exe
```

Внешних зависимостей нет, используется только стандартная библиотека Go.

### Использование

Запустите `steam-icon-repair.exe` (двойным кликом или из терминала).

```
steam-icon-repair.exe              # просканировать и починить
steam-icon-repair.exe -dry-run     # только показать сломанные, ничего не качать
steam-icon-repair.exe -v           # подробный вывод (включая целые иконки)
```

Пример успешного запуска:

```
Steam Icon Repair
=================
Steam shortcuts found: 14
Missing icons: 3

  appid=440  b32f4ac7d8e9f1c8.ico ... OK
  appid=730  a91c5e7d23b4f8a2.ico ... OK
  appid=570  c84d1e9b6f3a7d5e.ico ... OK

Restored: 3   Failed: 0

To refresh icons in Explorer, run:
  ie4uinit.exe -show
If icons still look stale, sign out and back in.
```

После восстановления иконок рекомендуется обновить кэш Windows Explorer командой `ie4uinit.exe -show`. Если это не помогает — выйдите из учётной записи и зайдите обратно.

### Что утилита НЕ делает

Эти ограничения сделаны осознанно, чтобы инструмент оставался простым и предсказуемым.

- **Не парсит `.lnk`-ярлыки.** Steam по умолчанию создаёт `.url`. Если вы вручную сделали `.lnk` — он будет пропущен.
- **Не чинит иконки внутри клиента Steam.** Утилита работает только с ярлыками. Если в самом клиенте Steam в библиотеке у игры пропала иконка — пара `(appid, hash)` неизвестна без ярлыка, и достать её можно только через Steam Web API (нужен ключ) или парсинг HTML магазина. Это за рамками задачи.
- **Не трогает кэш иконок Windows.** Файл `%LOCALAPPDATA%\IconCache.db` остаётся как есть — утилита только подсказывает, как его обновить.
- **Не модифицирует сами `.url`-файлы.** Только восстанавливает `.ico`-файл по уже прописанному в ярлыке пути.

### Возможные ошибки и их причины

| Сообщение | Причина |
|---|---|
| `HTTP 404` | Иконка с таким `hash` отсутствует на CDN. Чаще всего это значит, что ярлык создан для устаревшей версии иконки (Steam её заменил), или для пиратской/неофициальной сборки. Решение: пересоздайте ярлык через клиент Steam (правый клик по игре → «Управление» → «Добавить ярлык на рабочий стол»). |
| `HTTP 403` | Региональная блокировка CDN или временная проблема. Попробуйте через VPN. |
| `dial tcp: ...` | Нет интернета или CDN недоступен. |
| `No Steam shortcuts found.` | На обследованных рабочих столах и в меню «Пуск» нет ни одного `.url`-ярлыка, ссылающегося на `steam://rungameid/...`. |

### FAQ

**Безопасно ли это?**
Да. Утилита только читает `.url`-файлы и пишет `.ico`-файлы в их собственный исходный путь. Никаких изменений в реестре, в файлах Steam, в самих ярлыках не делается.

**Нужны ли права администратора?**
Обычно нет — иконки лежат в `Program Files (x86)\Steam\steam\games\`, но эта папка доступна для записи обычному пользователю на стандартных установках Steam. Если получаете `permission denied` — запустите утилиту от имени администратора.

**Почему `.ico`, а не `.png`?**
Windows для ярлыков использует именно `.ico` (контейнер с несколькими разрешениями в одном файле). Steam выкладывает их на CDN ровно в том же формате.

**Откуда Steam берёт хэш в имени файла?**
Это SHA1 содержимого `.ico`. Используется как cache-busting: разработчик загружает новую иконку → получает новый хэш → старая версия остаётся валидной для уже созданных ярлыков.

### Лицензия

MIT

---

## English

A tiny CLI utility for Windows that restores missing `.ico` icons of Steam game shortcuts on the desktop.

### Why it exists

When Steam creates a desktop shortcut for a game, it stores the `.ico` file at `<SteamPath>\steam\games\<hash>.ico`. This file can disappear for various reasons:

- antivirus removed a "suspicious" binary from a system folder;
- disk cleanup or a third-party "optimizer" wiped the cache;
- a bug during a Steam update;
- the user manually cleaned the `steam\games` folder.

The result: a game shortcut on the desktop with the default Windows icon. `Steam Icon Repair` finds such shortcuts and re-downloads the original icons from Valve's public CDN.

### How it works

Every Steam game shortcut is a `.url` file like this:

```ini
[InternetShortcut]
URL=steam://rungameid/440
IconFile=C:\Program Files (x86)\Steam\steam\games\b32f4...c8.ico
IconIndex=0
```

The utility extracts two values from this file:

- **`appid`** — from the `URL=steam://rungameid/<appid>` line;
- **`hash`** — the `.ico` filename without extension from the `IconFile=` line.

Given the `(appid, hash)` pair, it builds the public Steam CDN URL and downloads the icon directly:

```
https://cdn.cloudflare.steamstatic.com/steamcommunity/public/images/apps/<appid>/<hash>.ico
```

The file is saved at the original `IconFile=` path. No API keys required — these are public Steam Community assets.

### Where it scans for shortcuts

- `%USERPROFILE%\Desktop`
- `%PUBLIC%\Desktop`
- `%APPDATA%\Microsoft\Windows\Start Menu\Programs`
- `%PROGRAMDATA%\Microsoft\Windows\Start Menu\Programs`

Recursively, by `.url` extension.

### Installation

#### Option 1: prebuilt binary

Download `steam-icon-repair.exe` from the [Releases](../../releases) page and drop it anywhere.

#### Option 2: build from source

Requires Go 1.21+.

```bash
git clone https://github.com/<user>/steam-icon-repair.git
cd steam-icon-repair
go build -o steam-icon-repair.exe
```

Cross-compiling from macOS/Linux:

```bash
GOOS=windows GOARCH=amd64 go build -o steam-icon-repair.exe
```

No external dependencies — Go standard library only.

### Usage

Run `steam-icon-repair.exe` (double-click or from a terminal).

```
steam-icon-repair.exe              # scan and repair
steam-icon-repair.exe -dry-run     # only list broken icons, do not download
steam-icon-repair.exe -v           # verbose output (includes intact icons)
```

Example successful run:

```
Steam Icon Repair
=================
Steam shortcuts found: 14
Missing icons: 3

  appid=440  b32f4ac7d8e9f1c8.ico ... OK
  appid=730  a91c5e7d23b4f8a2.ico ... OK
  appid=570  c84d1e9b6f3a7d5e.ico ... OK

Restored: 3   Failed: 0

To refresh icons in Explorer, run:
  ie4uinit.exe -show
If icons still look stale, sign out and back in.
```

After restoring icons, refresh the Windows Explorer cache with `ie4uinit.exe -show`. If that doesn't help, sign out and back in.

### What it does NOT do

These limitations are intentional — the tool stays simple and predictable.

- **Does not parse `.lnk` shortcuts.** Steam creates `.url` by default. Manually created `.lnk` shortcuts are skipped.
- **Does not fix icons inside the Steam client.** The tool only works with desktop/Start Menu shortcuts. If a game's icon is missing inside the Steam library UI, the `(appid, hash)` pair is unknown without a shortcut — fetching it would require the Steam Web API (needs a key) or scraping the store page HTML. Out of scope.
- **Does not touch the Windows icon cache.** `%LOCALAPPDATA%\IconCache.db` is left alone — the tool only prints a hint on how to refresh it.
- **Does not modify the `.url` files themselves.** It only restores the `.ico` file at the path already specified in the shortcut.

### Possible errors

| Message | Cause |
|---|---|
| `HTTP 404` | No icon with that `hash` on the CDN. Usually means the shortcut was created against an older icon revision (Steam replaced it), or for a pirated/unofficial build. Fix: recreate the shortcut via the Steam client (right-click the game → Manage → Add desktop shortcut). |
| `HTTP 403` | Regional CDN block or transient issue. Try via VPN. |
| `dial tcp: ...` | No internet or CDN unreachable. |
| `No Steam shortcuts found.` | None of the scanned desktops or Start Menu folders contain `.url` shortcuts pointing to `steam://rungameid/...`. |

### FAQ

**Is it safe?**
Yes. The utility only reads `.url` files and writes `.ico` files to their own original paths. No registry changes, no Steam file modifications, no shortcut edits.

**Does it need admin rights?**
Usually not — icons live under `Program Files (x86)\Steam\steam\games\`, which is writable by the regular user in standard Steam installations. If you get `permission denied`, run as administrator.

**Why `.ico` and not `.png`?**
Windows shortcuts use `.ico` — a container holding multiple resolutions in a single file. Steam serves them on the CDN in the same format.

**Where does the hash in the filename come from?**
It's the SHA1 of the `.ico` contents. Used as a cache-buster: when a developer uploads a new icon, it gets a new hash, and old shortcuts referencing the old hash keep working until recreated.

### License

MIT
