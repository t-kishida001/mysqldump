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

### Edit Examples  

- .sql.cnf
```.sql.cnf
[client]
user = mysqlbackup
password = password
host = 192.168.1.11
port = 6446
```
DB接続に必要な情報を記載する。

`user`		: mysqldumpコマンドで指定する接続ユーザ  
`password`	: `user`で指定したユーザのパスワード  
`host`		: 接続アドレスorホスト名
`port`		: 接続ポート



- .env.cnf
```.env.cnf
DATABASES=DB1,DB2,mysql
DUMP_GENERATIONS=2
DUMP_DIR=/tmp/dump_directory
```
取得するデータベース、保持世代数など記載する。  

`DATABASES`		: dumpを取得するデータベース。カンマ区切りで複数指定可能  
`DUMP_GENERATIONS`	: 保持世代数。この指定を超えたdumpファイルは古いものから削除される  
`DUMP_DIR`		: dump格納パス。指定しなかった場合スクリプト実行ディレクトリに出力する  

DATABASESはカンマ区切りで複数指定可能

```bash
$ go run main.go

$ ls -l <DUMP_DIR>
```


## Requirement
- mysql: Ver 8.0.33-0ubuntu0.22.04.2 for Linux on x86_64 
- mysqldump:  Ver 8.0.33-0ubuntu0.22.04.2 for Linux on x86_64 ((Ubuntu))
- MySQL Router:  Ver 8.0.33 for Linux on x86_64 (MySQL Community - GPL)
- golang: version go1.13.8 linux/amd64 , version go1.20.6 linux/amd64
