# mysqldump

execute mysqldump

## Overview

Group Replication環境のmysqlデータベースにてmysqlrouterアドレス越しに取得するためのスクリプト。
確認していないが、mysql単体向けにも動作は可能(なはず)

実行環境からデータベースへ接続できること、以下に記すcnf中に設定するユーザがDBに作成されていること、あわせて権限が足りていることを前提とする。

## Usage
テンプレートファイルを編集し、リネームする

.sql.cnf.template > .sql.cnf
.env.txt.template > .env.txt

examples
```.sql.cnf
[client]
user = mysqlbackup
password = password
host = 192.168.0.1
port = 6446
```
DB接続に必要な情報を記載する。

```.env.cnf
DATABASES=test,test2,mysql
DUMP_GENERATIONS=2
DUMP_DIR=/tmp/dump_directory
```
取得するデータベース、保持世代数、dump保持ディレクトリパスを記載する。
DATABASESはカンマ区切りで複数指定可能

## Requirement
- mysql: Ver 8.0.33-0ubuntu0.22.04.2 for Linux on x86_64 
- mysqldump:  Ver 8.0.33-0ubuntu0.22.04.2 for Linux on x86_64 ((Ubuntu))
- MySQL Router:  Ver 8.0.33 for Linux on x86_64 (MySQL Community - GPL)
- golang: version go1.13.8 linux/amd64
