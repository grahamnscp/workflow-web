/* mysql init.sql */

CREATE USER 'mysqluser'@'%' IDENTIFIED WITH mysql_native_password BY 'mysqlpw';


/* Database for mock bank */
CREATE DATABASE dataentry;
GRANT ALL ON dataentry.* TO 'mysqluser'@'%';

/* Make sure the privileges are installed */
FLUSH PRIVILEGES;

USE dataentry;

DROP TABLE IF EXISTS accounts;
CREATE TABLE accounts (
  account_id  	  INT UNSIGNED NOT NULL AUTO_INCREMENT,
  account_number  INT NOT NULL,
  account_name 	  VARCHAR(30) NOT NULL,
  account_balance FLOAT NOT NULL,
  email           VARCHAR(40) NOT NULL,
  datestamp       TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (account_id),
  UNIQUE KEY transfer (account_name)
) ENGINE=INNODB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;

DROP TABLE IF EXISTS bankapistatus;
CREATE TABLE bankapistatus (
  up  INT NOT NULL DEFAULT 1
) ENGINE=INNODB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;

INSERT INTO bankapistatus (up) VALUES (1);

/* Database for moneytransfer mock queue */
CREATE DATABASE moneytransfer;
GRANT ALL ON moneytransfer.* TO 'mysqluser'@'%';

/* Make sure the privileges are installed */
FLUSH PRIVILEGES;

USE moneytransfer;

DROP TABLE IF EXISTS transfer;
CREATE TABLE transfer (
  id          INT UNSIGNED NOT NULL AUTO_INCREMENT,
  origin      VARCHAR(30) NOT NULL,
  destination VARCHAR(30) NOT NULL,
  amount      FLOAT NOT NULL,
  reference   VARCHAR(40) NOT NULL,
  status      VARCHAR(30) NOT NULL,
  t_wkfl_id   VARCHAR(50) NULL,
  t_run_id    VARCHAR(50) NULL,
  t_taskqueue VARCHAR(50) NULL,
  t_info      VARCHAR(250) NULL,
  datestamp   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY transfer (origin,destination,reference,amount,datestamp)
) ENGINE=INNODB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;

