# Analytics on GO

Анализ трансляции на GO 

## Instalation and run

```bash
git clone https://github.com/kirakashin/AnalyticsGO
go get 	github.com/gorilla/mux
go get	github.com/xuri/excelize/v2
cd AnalyticsGO/AnalyticsGO
go run main.go
```

## Usage

Проверяем работоспособность через /ping
```bash
curl --location --request GET 'localhost:8000/ping'
```

Проверяем колличество записей через /stat
```bash
curl --location --request GET 'localhost:8000/stat'
```

Посылаем отчет через /collect
```bash
curl --location --request POST 'localhost:8000/collect' \
--header 'Content-Type: application/json' \
--data-raw '[{
    "viewerId": "11181",
    "name": "Сергей",
    "lastName": "Сергеев",
    "isChatName": false,
    "email": "qwer@rewq.ru",
    "isChatEmail": false,
    "joinTime": "2021-07-30T14:12:48+03:00",
    "leaveTime": "2021-07-30T14:25:25+03:00",
    "spentTime": 676000000000,
    "spentTimeDeltaPercent": 9,
    "chatCommentsTotal": 0,
    "chatCommentsDeltaPercent": 0,
    "anotherFields": [],
    "browserClientInfo": {
        "userIP": "79.136.131.4",
        "platform": "Windows 10 64-bit",
        "browserClient": "Chrome 92.0.4515.107",
        "screenData_viewPort": "1920x1040",
        "screenData_resolution": "1920x1080"
    }
}]'
```

<!-- Получаем отчет по провайдерам/странам, разрешениям экранов, устройставам/браузерам (подробный) и пикам просмотров через /report_cp, /report_res, /report_os и /report_peaks соответственно
```bash
curl --location --request GET 'localhost:8000/report_cp'
curl --location --request GET 'localhost:8000/report_res'
curl --location --request GET 'localhost:8000/report_os'
curl --location --request GET 'localhost:8000/report_peaks'
``` -->

Общий красивый отчет можно получить через /report_all
```bash
curl --location --request GET 'localhost:8000/report_all'
```

Поменять порт можно в config.json