# OUT Language · The Go, but with a human face

![Version](https://img.shields.io/badge/version-0.4.0-blue)
![Go](https://img.shields.io/badge/Go-1.21+-00ADD8)
![License](https://img.shields.io/badge/license-MIT-green)
![Status](https://img.shields.io/badge/status-alpha-orange)

**OUT** — скриптовый язык программирования-обёртка над Go. Мы берём всю мощь Go и оборачиваем в дружелюбный, мемный и простой синтаксис. Это как Python с C++, только **OUT с Go**.

```out
print("Hello, OUT!")

def add(a, b) {
    return a + b
}

print(add(2, 3))                  // 5
print(strings::upper("out rocks")) // OUT ROCKS
print(math::random_int(1, 100))    // случайное число 1..100
vibe()                             // good vibes only
```

---

## 🚀 Почему OUT?

| Python + C++ | **OUT + Go** |
|--------------|--------------|
| Пишешь просто, под капотом мощь C | Пишешь просто, под капотом мощь Go |
| C-расширения для производительности | Go-модули для функциональности |
| Легко изучать новичку | Тот же дружелюбный подход |

Ты пишешь **дружелюбный скрипт**, а весь могучий Go-рантайм — горутины, пакеты, экосистема — работает прямо под капотом.

---

## ✨ Возможности

### Ядро языка
- Переменные, арифметика, строки, числа
- `if / else`, `while`, `for i in 1..5`
- Функции `def name(args) {}` и рекурсия
- Массивы `[1, 2, 3]` и словари `{"name": "Alex"}`
- Мем-функции: `vibe()`, `yeet()`, `bruh()`, `flex()`, `sus()`
- `import` между `.out` файлами

### Модули (обёртки над Go)
```
strings::   upper, lower, split, join, contains, replace, trim, repeat
math::      abs, sqrt, sin, cos, round, pow, log, random, pi
os::        cwd, mkdir, listdir, getenv
strconv::   atoi, itoa, parse_float, format_float
json::      parse, stringify
http::      get, post
files::     read, write, append, exists, list, mkdir, remove
crypto::    md5, sha1, sha256, base64_encode, base64_decode
time::      now, unix, millis, sleep, format
```

---

## ⚙️ Установка

Нужен **Go >= 1.21**.

```bash
# собрать интерпретатор
go build -o out.exe ./cmd/out/

# запустить REPL (живая консоль)
./out.exe

# запустить скрипт
./out.exe run examples/hello.out
```

---

## 🖥️ REPL (живая консоль)

Запусти `./out.exe` и сразу экспериментируй:

```text
>> 1 + 2
= 3

>> "a" + "b"
= ab

>> def square(n) { return n * n }
>> square(9)
= 81

>> [1, 2, 3].len
= 3

>> math::random_int(1, 100)
= 42
```

Введи `exit` чтобы выйти.

---

## 📦 Примеры

| Пример | Описание |
|--------|----------|
| `examples/hello.out` | Основы синтаксиса |
| `examples/modules.out` | Встроенные Go-модули |
| `examples/import_demo.out` | Импорт между файлами |
| `examples/utils.out` | Библиотека переиспользуемых функций |

```bash
out.exe run examples/hello.out
out.exe run examples/modules.out
```

---

## 🗂️ Структура проекта

```
out-lang/
├── cmd/out/            # CLI и REPL (точка входа)
├── internal/
│   ├── lexer/          # токенизация исходного кода
│   ├── parser/         # построение AST
│   ├── ast/            # узлы синтаксического дерева
│   ├── eval/           # интерпретатор (обход AST)
│   ├── module/         # ядро системы модулей (Go-расширения)
│   ├── object/         # система типов
│   └── stdlib/         # стандартная библиотека (обёртки над Go)
└── examples/           # примеры скриптов
```

---

## 🏛️ Архитектура

```
┌─────────────────────────────────────────────┐
│             OUT-код (скрипты)               │
│        import "file.out", def, loops       │
└─────────────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────────┐
│         Система модулей (module)            │
│      регистрация и вызов Go-модулей        │
└─────────────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────────┐
│     Стандартная библиотека (stdlib)        │
│     обёртки над Go: strings, math, http…   │
└─────────────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────────┐
│           Go-рантайм (весь Go)             │
│    горутины, GC, пакеты, экосистема       │
└─────────────────────────────────────────────┘
```

---

## 🗺️ Дорожная карта

- [x] **v0.4** — ядро: лексер, парсер, интерпретатор, REPL
- [x] **v0.4+** — модульная система и стандартная библиотека
- [ ] **v0.5** — обработка ошибок (`try / catch`, `?` оператор)
- [ ] **v0.6** — расширенные коллекции и методы
- [ ] **v0.7** — менеджер пакетов
- [ ] **v1.0** — стабильный релиз + CLI-инструменты

---

## 📄 Лицензия

MIT
