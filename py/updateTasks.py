import sqlite3

database_name="d:\\GoSpace\\src\\greyman\\dicton\\data.sqlite3"
con = sqlite3.connect(database_name)
cursor=con.cursor()

f=open("d:\\exist.txt","r",encoding="utf-8")


for rec in f:
    cursor.execute("insert into tasks values (null,?,?,?,0)", rec.rstrip().split("\t"))

con.commit()
con.close()