# OUT Language

**OUT** — это язык программирования-обёртка над Go. Мы берём всю мощь Go и оборачиваем в дружелюбный, мемный, скриптовый синтаксис. Это как Python с C++, только OUT — с Go.

```
print("Hello, OUT!")
def add(a, b) { return a + b }
print(add(2, 3))   // 5

print(strings::upper("out rocks"))  // OUT ROCKS
print(math::random_int(1, 100))     // случайное число
```

## Возможности

- Переменные, арифметика, строки
- `if / else`, `while`, `for i in 1..5`
- Функции `def name(args) {}` и рекурсия
- Массивы `[1, 2, 3]` и словари `{"name": "Alex"}`
- Мем-функции: `vibe()`, `yeet()`, `bruh()`, `flex()`, `sus()`
- `import` между `.out` файлами
- Модульная система поверх Go: `strings::`, `math::`, `os::`, `json::`, `http::`, `files::`, `crypto::`, `time::`, `list::`, `strconv::`

## Установка

Нужен Go (>= 1.21):

```bash
# собрать интерпретатор
go build -o out.exe ./cmd/out/

# REPL
./out.exe

# запуск скрипта
./out.exe run examples/hello.out
```

## Примеры

```
C:\...\out-lang> out.exe run examples/hello.out
C:\...\out-lang> out.exe run examples/modules.out
```

## Структура проекта

```
out-lang/
├── cmd/out/            # CLI и REPL
├── internal/
│   ├── lexer/          # токенизация
│   ├── parser/         # построение AST
│   ├── ast/            # узлы дерева
│   ├── eval/           # интерпретатор
│   ├── module/         # ядро системы модулей
│   ├── object/         # система типов
│   └── stdlib/         # стандартная библиотека (обёртки над Go)
└── examples/           # примеры скриптов
```

## Схема архитектуры

```
OUT-код (скрипты)
    │
    ▼
Система модулей (internal/module)
    │
    ▼
Стандартная библиотека (stdlib) — обёртки над Go-пакетами
    │
    ▼
Go-рантайм (весь Go: горутины, пакеты, экосистема)
```
