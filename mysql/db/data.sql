/* data.sql */

/* database dataentry */
insert into dataentry.accounts (account_number,account_name,account_balance,email) values (1001,"jane",110.00,"jane@mytelco.io");
insert into dataentry.accounts (account_number,account_name,account_balance,email) values (1002,"bill",1000.00,"billy@bob.net");
insert into dataentry.accounts (account_number,account_name,account_balance,email) values (1003,"ted",10.00,"ted10@gym.org");
insert into dataentry.accounts (account_number,account_name,account_balance,email) values (1004,"sally",1000.00,"sals@petnet.com");
insert into dataentry.accounts (account_number,account_name,account_balance,email) values (1005,"harry",1000.00,"harryk@us.now");
insert into dataentry.accounts (account_number,account_name,account_balance,email) values (1006,"jim",1000.00,"jim.burns@bt.internet");
insert into dataentry.accounts (account_number,account_name,account_balance,email) values (1007,"rich",20000.00,"rich@lom.net");

/* database moneytransfer */
insert into moneytransfer.transfer (origin,destination,amount,reference,status) values ('bill','jim',120,'IOU','REQUESTED');
insert into moneytransfer.transfer (origin,destination,amount,reference,status) values ('jane','sally',107,'FOOD MONEY','REQUESTED');
insert into moneytransfer.transfer (origin,destination,amount,reference,status) values ('ted','harry',100,'CART123','REQUESTED');
