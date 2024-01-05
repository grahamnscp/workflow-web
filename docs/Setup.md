# workflow-web component setup

## Temporal Cloud configuration
This example assumes that you have a temporal cloud configured and have local client certificate files for your namespace.
The values are passed into the demo app using environment variables, example direnv .envrc file is included in the repo

Source or cp .envrc.example to .envrc and `direnv allow`:
[.envrc.example](../.envrc.example)
  
## Docker component dependancies
The various modules of the webapp framework use a number of components that run in docker.  

- mysql
- mailhog
- mongodb

Ensure that docker is installed and running on your system.  


### Start local mysql database
The sample mysql database has been configured to run using docker-compose locally and initialise the database with users and sample data.
```
cd mysql
docker-compose up -d
docker-compose ps
NAME      IMAGE          COMMAND                  SERVICE   CREATED      STATUS      PORTS
mysql     mysql:latest   "docker-entrypoint.s…"   mysql     2 days ago   Up 2 days   0.0.0.0:3306->3306/tcp, 33060/tcp
```

Sample data:
```
docker exec -it mysql mysql -u root -p
dbroot

Welcome to the MySQL monitor.  Commands end with ; or \g.
...

mysql> use dataentry;
Database changed

mysql> show tables;
+---------------------+
| Tables_in_dataentry |
+---------------------+
| accounts            |
| bankapistatus       |
+---------------------+
2 rows in set (0.00 sec)

mysql> describe accounts;
+-----------------+--------------+------+-----+-------------------+-------------------+
| Field           | Type         | Null | Key | Default           | Extra             |
+-----------------+--------------+------+-----+-------------------+-------------------+
| account_id      | int unsigned | NO   | PRI | NULL              | auto_increment    |
| account_number  | int          | NO   |     | NULL              |                   |
| account_name    | varchar(30)  | NO   | UNI | NULL              |                   |
| account_balance | float        | NO   |     | NULL              |                   |
| email           | varchar(40)  | NO   |     | NULL              |                   |
| datestamp       | timestamp    | NO   |     | CURRENT_TIMESTAMP | DEFAULT_GENERATED |
+-----------------+--------------+------+-----+-------------------+-------------------+
6 rows in set (0.01 sec)

mysql> select * from accounts;
+------------+----------------+--------------+-----------------+-----------------------+---------------------+
| account_id | account_number | account_name | account_balance | email                 | datestamp           |
+------------+----------------+--------------+-----------------+-----------------------+---------------------+
|          1 |           1001 | jane         |             110 | jane@mytelco.io       | 2024-01-03 07:59:33 |
|          2 |           1002 | bill         |            1000 | billy@bob.net         | 2024-01-03 07:59:33 |
|          3 |           1003 | ted          |              10 | ted10@gym.org         | 2024-01-03 07:59:33 |
|          4 |           1004 | sally        |            1000 | sals@petnet.com       | 2024-01-03 07:59:33 |
|          5 |           1005 | harry        |            1000 | harryk@us.now         | 2024-01-03 07:59:33 |
|          6 |           1006 | jim          |            1000 | jim.burns@bt.internet | 2024-01-03 07:59:33 |
|          7 |           1007 | rich         |           20000 | rich@lom.net          | 2024-01-03 07:59:33 |
+------------+----------------+--------------+-----------------+-----------------------+---------------------+
11 rows in set (0.00 sec)

mysql> use moneytransfer;
Database changed

mysql> show tables;
+-------------------------+
| Tables_in_moneytransfer |
+-------------------------+
| transfer                |
+-------------------------+
1 row in set (0.03 sec)

mysql> describe moneytransfer.transfer;
+-------------+--------------+------+-----+-------------------+-------------------+
| Field       | Type         | Null | Key | Default           | Extra             |
+-------------+--------------+------+-----+-------------------+-------------------+
| id          | int unsigned | NO   | PRI | NULL              | auto_increment    |
| origin      | varchar(30)  | NO   | MUL | NULL              |                   |
| destination | varchar(30)  | NO   |     | NULL              |                   |
| amount      | float        | NO   |     | NULL              |                   |
| reference   | varchar(40)  | NO   |     | NULL              |                   |
| status      | varchar(30)  | NO   |     | NULL              |                   |
| t_wkfl_id   | varchar(50)  | YES  |     | NULL              |                   |
| t_run_id    | varchar(50)  | YES  |     | NULL              |                   |
| t_taskqueue | varchar(50)  | YES  |     | NULL              |                   |
| t_info      | varchar(250) | YES  |     | NULL              |                   |
| datestamp   | timestamp    | NO   |     | CURRENT_TIMESTAMP | DEFAULT_GENERATED |
+-------------+--------------+------+-----+-------------------+-------------------+
11 rows in set (0.00 sec)

mysql> select id,origin,destination,amount,reference,status from moneytransfer.transfer;
+----+--------+-------------+--------+------------------+-----------+
| id | origin | destination | amount | reference        | status    |
+----+--------+-------------+--------+------------------+-----------+
|  1 | bill   | jim         |    120 | IOU              | REQUESTED |
|  2 | jane   | sally       |    107 | FOOD MONEY       | REQUESTED |
|  3 | ted    | harry       |    100 | CART123          | REQUESTED |
|  4 | bill   | ted         |     10 | transfer request | REQUESTED |
+----+--------+-------------+--------+------------------+-----------+
4 rows in set (0.00 sec)

mysql> quit
Bye
```


### Start local mailhog service
Some of the demo apps use email notifications with call back actions, mailhog docker image is used for this
```
cd mailhog
docker-compose up -d
docker-compose ps
NAME      IMAGE                    COMMAND     SERVICE   CREATED        STATUS      PORTS
mailhog   mailhog/mailhog:v1.0.1   "MailHog"   mailhog   6 months ago   Up 7 days   0.0.0.0:1025->1025/tcp, 0.0.0.0:8025->8025/tcp
```
The mailhog UI is available at [localhost:8025](http://localhost:8025)



### Start local mongodb service
The new account onboarding workflow uses various mongodb collections
```
cd mongodb
docker-compose up -d
docker-compose ps
mongodb   mongo:latest    "docker-entrypoint.s…"   mongodb         2 days ago   Up 2 days   0.0.0.0:27017->27017/tcp
mongoui   mongo-express   "/sbin/tini -- /dock…"   mongo-express   2 days ago   Up 2 days   0.0.0.0:8081->8081/tcp
```
The mongodb UI is available at [localhost:8081](http://localhost:8081)



## Custom Search Attribute
The demo app makes use of a custom search attribute that needs to be created in Temporal Cloud for the namespace:   

Name: CustomStringField, Type: Text



## Temporal Cloud Web UI - Data Converter

If you are encrypting the workflow payload content as well then to inspect the workflow history data values you can connect to a localhost codec server that implements /decode url configured for the same dataconverter used in your temporal client app code.   

Start the local codec server (in a new terminal window) using:
```
cd dataconverter/codec-server-go
go run server.go

2023/06/06 11:48:46 Serve Http on 8888
```

