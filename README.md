# OUT Language · The Go, but with a human face

![Version](https://img.shields.io/badge/version-0.5.0-blue)
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

# обработка ошибок
try {
    throw "что-то пошло не так"
} catch(e) {
    print("Ошибка: " + e)
}

# безопасный доступ
user = {name: "Bob"}
print(user?.name)  // Bob
print(user?.city)  // null (без паники)
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
- Массивы `[1, 2, 3]` и словари `{name: "Bob"}`
- `import` между `.out` файлами
- **try / catch / throw** — обработка ошибок
- **`?.`** — безопасный доступ (возвращает `null` вместо паники)
- **`#`** — комментарии (как в Python)

### Команды CLI
```
out                     # REPL
out run <file.out>      # запуск скрипта
out compile <file.out>  # компиляция в standalone .exe
out get <library>       # скачивание библиотеки с GitHub
out libs                # список установленных библиотек
out errors <file.out>   # показ ошибок компиляции
```

### Модули (обёртки над Go)
```
strings::   upper, lower, split, join, contains, replace, trim
math::      abs, sqrt, sin, cos, round, pow, log, pi
random::    int, float, choice, shuffle, seed
os::        env, args, system
json::      parse, stringify
http::      get, post
files::     read, write, exists, remove, list, mkdir
crypto::    md5, sha1, sha256, encrypt, decrypt
time::      now, format, timestamp, sleep
array::     filter, map, reduce, sort, unique, find, any, all
dict::      keys, values, merge, get
logging::   debug, info, warn, error
console::   clear, color, size
dev::       board, pinMode, digitalWrite, analogRead
```

### OUT IDE
- Подсветка синтаксиса
- Автосохранение
- Компиляция в .exe (Ctrl+B)
- Проверка ошибок (Ctrl+T)
- Встроенные библиотеки (панель справа)
- Копирование ошибок (Ctrl+E)

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

# скомпилировать в standalone .exe
./out.exe compile myapp.out myapp.exe
```

---

## 📦 Примеры

| Пример | Описание |
|--------|----------|
| `examples/hello.out` | Основы синтаксиса |
| `examples/modules.out` | Встроенные модули |
| `examples/import_demo.out` | Импорт между файлами |

```bash
out run examples/hello.out
out run examples/modules.out
```

---

## 🗂️ Структура проекта

```
out-lang/
├── cmd/out/            # CLI и REPL (точка входа)
├── internal/
│   ├── lexer/          # токенизация (+ # комментарии)
│   ├── parser/         # построение AST (+ try/catch/throw, ?)
│   ├── ast/            # узлы синтаксического дерева
│   ├── eval/           # интерпретатор (+ try/catch/throw, ?)
│   ├── env/            # области видимости
│   ├── module/         # ядро системы модулей
│   ├── object/         # система типов
│   ├── libs/           # менеджер библиотек (out get)
│   └── stdlib/         # стандартная библиотека (18 модулей)
├── libs/               # встроенные .out библиотеки
└── examples/           # примеры скриптов
```

---

## 🗺️ Дорожная карта

- [x] **v0.4** — ядро: лексер, парсер, интерпретатор, REPL
- [x] **v0.4+** — модульная система и стандартная библиотека
- [x] **v0.5** — обработка ошибок (`try / catch`, `?` оператор)
- [ ] **v0.6** — расширенные коллекции и методы
- [ ] **v0.7** — менеджер пакетов
- [ ] **v1.0** — стабильный релиз + CLI-инструменты

---

## 📄 Лицензия

MIT
