# Кроссплатформенный наблюдатель за изменениями файлов

## Данное приложение позволяет:

- Рекурснивно следить за изменениями в различных директориях
- Выполнять произвольный набор консольных команд

### Для запуска приложения:

1. Создать Базу Данных и указать в файле конфигурации "Config.yaml" значения самой БД (по умолчанию значения уже стоят в файле).
2. Создать файл ".env", где указать пароль БД с ключом DB_PASSWORD (пример: DB_PASSWORD=qwerty).
3. БД можно развернуть в Docker и команда будет выглядить следующим образом (значения стоят по умолчанию):

```
docker run --name=observer-db -e POSTGRES_PASSWORD=qwerty -p 5438:5432 -d postgres
```

4. Нужно сделать миграции при первом запуске (значения стоят по умолчанию):

```
migrate -path ./schema -database "postgres://postgres:qwerty@localhost:5438/postgres?sslmode=disable" up
```

5. Теперь почти все готово, необходимо только указать некоторые значения для наблюдателя в файле конфигурации:
   - Значние "path" позволяет выбрать директорию для наблюдения
   - Значние "log_file" позволяет задать файл для логирования, куда будут отправляться логи от выполнения команд (без значения - выключено)
   - Значния "include_regexp" и "exclude_regexp" позволяют включать/отключать файл по маске
   - Значние "commands" позволяет выбрать команды, которые будут выполняться при изменении файлов (без значений - выключено)

### Вопросы, которые возникли при реализации данного ТЗ:

- Не было указано под какую ОС делать наблюдателя, потому что от этого зависит сама реализация. Была выбрана библиотека наблюдателя "fsnotify". Она позволяет сделать кроссплатформенного наблюдателя, поэтому он видит эвенты: Create, Remove, Rename, Write, and Chmod.
- В текущей версии приложения пути определяются с запуском программы. Если в наблюдаемом каталоге создать папку, то наблюдатель не сможет увидеть изменяемые файлы в этой папке. Наверное, стоит это исправить.
