# Смена пароля базы данных

## MySQL

Если ваша база данных работает в Docker-контейнере, войдите в неё, выполнив:

```bash
```
docker exec -it <container_id> bash

Войдите в CLI MySQL, выполнив:

```bash
```
mysql -u<root*mysql*user> -p<old_password>

Чтобы сменить пароль для MySQL 5.7 и новее, выполните:

```bash
```
mysql> SET PASSWORD FOR 'root' = PASSWORD('new_password');
mysql> FLUSH PRIVILEGES;

Чтобы сменить пароль для более старых версий (до 5.7), выполните:

```bash
```
mysql> ALTER USER '<mysql*user>'@'localhost' IDENTIFIED BY '<mysql*password>';
mysql> FLUSH PRIVILEGES;

## PostgreSQL

Если ваша база данных работает в Docker-контейнере, войдите в неё, выполнив:

```bash
```
docker exec -it <container_id> bash

Оказавшись внутри контейнера, смените пароль с помощью:

```bash
```
psql --user <postgresql_user>
\password
\q
