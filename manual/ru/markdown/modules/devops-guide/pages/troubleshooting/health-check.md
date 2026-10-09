# Проверка работоспособности (Health Check)

Проверка работоспособности — это ряд шагов, которые система выполняет для определения того, работает ли инстанс LowCoooode как задумано.

Чтобы получить доступ к автоматизированной проверке работоспособности, вы можете перейти на эндпоинт `/healthcheck` вашего инстанса LowCoooode.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "troubleshooting/health-check/http.png",
    "alias": "troubleshooting-health-check-http",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 0,
    "y": 0,
    "w": 204,
    "h": 109
  }
}

Кроме того, вы можете выполнить команду `docker-compose ps`, чтобы проверить работоспособность из CLI.

```
```
       Name                      Command                  State                Ports
---------------------------------------------------------------------------------------------
my-lowcode*db*1       /docker-entrypoint.sh mysqld     Up (healthy)   3306/tcp, 33060/tcp
my-lowcode*server*1   ./bin/lowcode-server serve-api   Up (healthy)   127.0.0.1:18080->80/tcp

Обратитесь к документации [Troubleshooting](modules/devops-guide/pages/troubleshooting/troubleshooting/index.md) за дополнительной помощью по исправлению неработоспособных частей системы.
